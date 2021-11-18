package core

import (
	secRnd "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// World represents the game board and contains all game elements (clouds).
// Other components such as server or GUI access the world via a reference and influence it.
// All exported methods in the world are thread safe.
type World struct {
	width     int // default 2048
	height    int // default 1024
	gameSpeed int // how often per second will the server update (DEFAULT: 60)

	// stats
	iteration    uint64
	worldVapor   float32
	alive        int
	winCondition bool
	leader       string

	// cloud list
	clouds []*Cloud
	mux    *sync.Mutex

	SimSpeedUp int  // dirty hack for faster simulations (DEFAULT: 1)
	freeze     bool // block updates (DEFAULT: false)
}

// NewWorld create a new World.
// width and height defines the game board dimensions. [DEFAULT: 2048 x 1152]
// gameSpeed tells other components how often the server will update the world per second (with Update()). [DEFAULT: 60]
// amount influences how many neutral clouds are generated (it may be less due to collisions). [DEFAULT: 100]
// initWind indicates how fast neutral clouds can be at maximum. [DEFAULT: 7]
// initSize indicates the maximum size of neutral clouds. [DEFAULT: 200]
func NewWorld(width, height, gameSpeed, amount int, initWind, initSize float32, seed int64) *World {
	// new world
	w := &World{
		width:      width,
		height:     height,
		gameSpeed:  gameSpeed,
		clouds:     make([]*Cloud, 0),
		mux:        new(sync.Mutex),
		SimSpeedUp: 1,
	}

	// generate random Clouds
	rnd := rand.New(rand.NewSource(seed))
	for i := 0; i < amount; i++ {
		pos := NewPosition(rnd.Float32()*float32(width), rnd.Float32()*float32(height))
		vel := NewVelocity((2*rnd.Float32()-1)*initWind, (2*rnd.Float32()-1)*initWind)
		vap := rnd.Float32() * initSize
		c := NewCloud(w, pos, vel, vap, "", "")
		w.addCloud(c)
	}

	// return
	return w
}

//----  GETTER  ------------------------------------------------------------------------------------------------------//

// Width is the game board width
func (w *World) Width() int {
	return w.width
}

// Height is the game board height
func (w *World) Height() int {
	return w.height
}

// GameSpeed means Updates() per second
func (w *World) GameSpeed() int {
	return w.gameSpeed
}

// MaxIterations returns the last interaction to trigger the win conditions (timeout: after 3 minutes)
func (w *World) MaxIterations() uint64 {
	return 3 * 60 * uint64(w.GameSpeed())
}

// Stats returns interesting world statistics.
// iteration is the current game round (increases with every update).
// worldVapor is the worldwide vapor.
// alive shows how many objects there are in the world.
func (w *World) Stats() (iteration uint64, worldVapor float32, alive int, winCondition bool, leader string) {
	w.mux.Lock()
	defer w.mux.Unlock()

	iteration = w.iteration
	worldVapor = w.worldVapor
	alive = w.alive
	winCondition = w.winCondition
	leader = w.leader
	return
}

// Clouds returns a list of clouds.
// This function is equal to clone() and creates a new instance
// of the list and initializes all its fields with exactly the contents.
func (w *World) Clouds() []*Cloud {
	w.mux.Lock()
	defer w.mux.Unlock()

	var ret = make([]*Cloud, 0, len(w.clouds))
	for _, c := range w.clouds {
		ret = append(ret, c.clone())
	}
	return ret
}

// Me finds the player cloud reference.
// Changes to the returned object affect the world.
// If there are several clouds with the same name, the first result (the oldest cloud) is always returned
func (w *World) Me(name string) *Cloud {
	w.mux.Lock()
	defer w.mux.Unlock()

	for _, c := range w.clouds {
		if c.Player == name {
			return c // return first result
		}
	}
	return nil // player cloud not found
}

// Clone creates a new instance of World and initializes all its fields with exactly the contents.
func (w *World) Clone() *World {
	w.mux.Lock()
	defer w.mux.Unlock()

	// new world
	ret := &World{
		width:     w.width,
		height:    w.height,
		gameSpeed: w.gameSpeed,

		iteration:    w.iteration,
		worldVapor:   w.worldVapor,
		alive:        w.alive,
		winCondition: w.winCondition,
		leader:       w.leader,

		clouds: make([]*Cloud, 0, len(w.clouds)),
		mux:    new(sync.Mutex),

		SimSpeedUp: w.SimSpeedUp,
		freeze:     w.freeze,
	}

	// set clouds
	for _, c := range w.clouds {
		ret.clouds = append(ret.clouds, c.clone())
	}

	// rep. world links
	for _, c := range ret.clouds {
		c.world = ret
	}

	// return
	return ret
}

//----  SETTER  ------------------------------------------------------------------------------------------------------//

// AddPlayer add a new player cloud to the world.
// If pos is nil, then a random position is chosen.
func (w *World) AddPlayer(name, color string, pos *Position, vapor float32) *Cloud {
	w.mux.Lock()
	defer w.mux.Unlock()

	// random position
	if pos == nil {
		rnd := rand.New(rand.NewSource(12345 + time.Now().UnixMicro()))
		for {
			// random position
			x := rnd.Float32() * float32(w.Width())
			y := rnd.Float32() * float32(w.Height())
			pos = NewPosition(x, y)
			// check other clouds
			tryAgain := false
			testC := NewCloud(w, pos, nil, vapor, "", "")
			for _, c := range w.clouds {
				if c.isIntersects(testC) {
					tryAgain = true
					break
				}
			}
			// success
			if !tryAgain {
				break
			}
		}
	}

	// add player
	c := NewCloud(w, pos, NewVelocity(0, 0), vapor, name, color)
	w.addCloud(c)
	return c
}

// Freeze can freeze the world. Then there are no updates and all movement commands are discarded.
func (w *World) Freeze(b bool) {
	w.mux.Lock()
	defer w.mux.Unlock()

	w.freeze = b
}

// Move executes the move command of the cloud in the world.
func (w *World) Move(c *Cloud, wind *Velocity) bool {
	w.mux.Lock()
	defer w.mux.Unlock()

	if c == nil || c.world == nil || c.world.freeze {
		return false
	} else {
		return c.move(wind)
	}
}

// Kill is a suicide order. The cloud explodes.
func (w *World) Kill(c *Cloud) bool {
	w.mux.Lock()
	defer w.mux.Unlock()

	if c != nil && c.world != nil {
		return c.kill()
	} else {
		return false
	}
}

// Update calls Cloud.update() for each cloud in the world.
// The world statistics are also calculated.
func (w *World) Update() {
	w.mux.Lock()
	defer w.mux.Unlock()

	// freeze
	if w.freeze {
		return
	}

	// update AND remove dead Clouds
	var worldVapor float32
	var alive int
	var newList = make([]*Cloud, 0, len(w.clouds))
	for _, c := range w.clouds {
		// changes
		c.update()

		// add to new list
		if !c.IsDeath() || c.Player != "" {
			newList = append(newList, c)
			worldVapor += c.Vapor
		}

		// stats
		if !c.IsDeath() {
			// count Alive
			alive++
		}
	}

	// set new attributes
	w.clouds = newList
	w.iteration++
	w.worldVapor = worldVapor
	w.alive = alive
	w.winCondition, w.leader = w.isWinner()
}

// addCloud is a helper (not thread-safe)
func (w *World) addCloud(c *Cloud) {
	// generate uid
	uid := make([]byte, 6)
	_, _ = secRnd.Read(uid)

	// set uid
	c.UID = base64.StdEncoding.EncodeToString(uid)

	// add to list
	w.clouds = append(w.clouds, c)
}

// isWinner returns whether the victory conditions have been met and who is currently in the lead.
// The world statistics are also calculated.
func (w *World) isWinner() (is bool, winner string) {
	// get best player
	var best *Cloud
	for _, c := range w.clouds {
		if c.Player != "" && !c.IsDeath() {
			if best == nil {
				best = c // set first player
			} else {
				if best.Vapor < c.Vapor {
					best = c // set better player
				}
			}
		}
	}

	// no player, no winner
	if best == nil {
		return true, "no player alive"
	}

	// timeout
	if w.iteration > w.MaxIterations() {
		return true, best.Player
	}

	// > 50 %
	if best.Vapor/w.worldVapor*100 > 51 {
		return true, best.Player
	}

	// default
	return false, best.Player
}

//----  Serialisation  -----------------------------------------------------------------------------------------------//

// define JsonWorld (export hidden vars)
type jsonWorld struct {
	Width        int
	Height       int
	GameSpeed    int
	Iteration    uint64
	WorldVapor   float32
	Alive        int
	WinCondition bool
	Leader       string
	Clouds       []*Cloud
	SimSpeedUp   int
}

// ToJson return the world as json string.
func (w *World) ToJson() string {
	w.mux.Lock()
	defer w.mux.Unlock()

	// init JsonWorld
	ret := &jsonWorld{
		Width:        w.width,
		Height:       w.height,
		GameSpeed:    w.gameSpeed,
		Iteration:    w.iteration,
		WorldVapor:   w.worldVapor,
		Alive:        w.alive,
		WinCondition: w.winCondition,
		Leader:       w.leader,
		Clouds:       w.clouds,
		SimSpeedUp:   w.SimSpeedUp,
	}

	// serialisation
	b, err := json.Marshal(ret)
	if err != nil {
		fmt.Printf("ERROR: ToJson: %v\n", err)
	}

	// return
	return string(b)
}

// FromJson override this world with a json world string.
func (w *World) FromJson(str string) {
	// fix mux, if world is empty
	if w.mux == nil {
		w.mux = new(sync.Mutex)
	}

	// lock
	w.mux.Lock()
	defer w.mux.Unlock()

	// de-serialisation
	jw := new(jsonWorld)
	if err := json.Unmarshal([]byte(str), jw); err != nil {
		fmt.Printf("ERROR: FromJson: %v\n", err)
	}

	// set new world
	w.width = jw.Width
	w.height = jw.Height
	w.gameSpeed = jw.GameSpeed
	w.iteration = jw.Iteration
	w.worldVapor = jw.WorldVapor
	w.alive = jw.Alive
	w.winCondition = jw.WinCondition
	w.leader = jw.Leader
	w.clouds = jw.Clouds
	w.SimSpeedUp = jw.SimSpeedUp

	// repair world links
	for _, c := range w.clouds {
		c.world = w
	}
}
