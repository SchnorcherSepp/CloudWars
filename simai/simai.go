package simai

import (
	"CloudWars/core"
	"CloudWars/remote"
	"math/rand"
	"strings"
	"time"
)

// RunSimAI is a simple AI that accesses a game server with
// the remote.TcpClient and controls a player cloud.
// The AI calculates the future (8 sec) of 64 random
// movement commands every 100 ms and executes the best option.
func RunSimAI(host, port, name, color string) {
	// init client
	t := remote.NewTcpClient(host, port)

	// set name and start game
	t.Name(name)
	t.Color(color)
	resp := t.Play()

	// check resp
	if !strings.HasPrefix(resp, "ok") {
		println(resp)
		return
	}

	// prepare move list
	windList := make([]*core.Velocity, 0)
	windList = append(windList, core.NewVelocity(0, 0))
	for i := 0; i < 360; i += 11 {
		windList = append(windList, core.NewVelocityByAngle(float32(i), 10))
	}
	for i := 0; i < 360; i += 11 {
		windList = append(windList, core.NewVelocityByAngle(float32(i), 50))
	}
	rand.Shuffle(len(windList), func(i, j int) { windList[i], windList[j] = windList[j], windList[i] })

	// ai loop
	for {
		var bestVapor = float32(0)
		var bestMove = core.NewVelocity(0, 0)
		var world = &core.World{}
		var dead = false
		world.FromJson(t.List())
		world.SimSpeedUp = 15

		// simulate
		for _, wind := range windList {
			// reset world
			w := world.Clone()
			// get player
			me := w.Me(name)
			if me.IsDeath() {
				dead = true
			}
			// simulate
			w.Move(me, wind)
			for i := 0; i < 8*world.GameSpeed()/world.SimSpeedUp; i++ { // check 8 seconds
				w.Update()
			}
			if bestVapor < me.Vapor {
				bestVapor = me.Vapor
				bestMove = wind
			}
		}

		// send move
		if dead {
			break
		}
		t.Move(bestMove)
		time.Sleep(50 * time.Millisecond) // Don't DOS the server
	}
	println("END", name)
	t.Close() //bye
}
