package remote

import (
	"CloudWars/core"
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
)

type server struct {
	host           string
	port           string
	initPlayerSize float32
	world          *core.World
	conn           *net.Conn
	waitPlayer     int
	players        []string
	mux            *sync.Mutex
}

// RunServer starts a server and makes the game world available remotely.
// The server controls remote player clouds the game via the world reference.
// The initPlayerSize attribute determines how much vapor remotely generated player clouds will have.
// The waitPlayer attribute controls how many players will be waited for.
func RunServer(host, port string, initPlayerSize float32, world *core.World, waitPlayer int) {

	// Listen for incoming connections.
	l, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("RunServer: %v\n", err)
	}

	// Close the listener when the application closes.
	defer func(l net.Listener) {
		_ = l.Close()
	}(l)

	// Freeze world
	world.Freeze(true) // undo in registerPlayer()

	// server
	ser := &server{
		host:           host,
		port:           port,
		initPlayerSize: initPlayerSize,
		world:          world,
		waitPlayer:     waitPlayer,
		players:        make([]string, 0, waitPlayer),
		mux:            new(sync.Mutex),
	}

	fmt.Println("START SERVER [" + host + ":" + port + "]")
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn, ser)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, ser *server) {

	// prepare line reader
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	// close at end
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	// vars
	var name = fmt.Sprintf("unknown [%s]", conn.RemoteAddr())
	var color = "red"
	var me *core.Cloud

	// loop
	for {
		// read one line (ended with \n or \r\n)
		line, _ := tp.ReadLine()

		// extract command
		var com string
		if len(line) >= 4 {
			com = strings.ToLower(line[:4])
		}

		// CHECK COMMANDS
		if com == "quit" || com == "exit" { //--------------------------------------------------------------------< EXIT
			comWrite(conn, "ok")
			break // exit loop and close connection

		} else if com == "list" { //------------------------------------------------------------------------------< LIST
			if comWrite(conn, ser.world.ToJson()) {
				break // exit loop and close connection
			}

		} else if com == "play" { //------------------------------------------------------------------------------< PLAY
			if me == nil {
				if e := ser.registerPlayer(name); e == nil {
					me = ser.world.AddPlayer(name, color, nil, ser.initPlayerSize)
					if comWrite(conn, "ok: the game begins when all players are ready") {
						break // exit loop and close connection
					}
				} else {
					if comWrite(conn, fmt.Sprintf("err: %v", e)) {
						break // exit loop and close connection
					}
				}
			} else {
				if comWrite(conn, "err: you're already playing") {
					break // exit loop and close connection
				}
			}

		} else if com == "kill" { //------------------------------------------------------------------------------< KILL
			if me == nil {
				if comWrite(conn, "err: you're not playing") {
					break // exit loop and close connection
				}
			} else {
				if me.IsDeath() {
					if comWrite(conn, "err: you're already dead") {
						break // exit loop and close connection
					}
				} else {
					ser.world.Kill(me)
					if comWrite(conn, "ok") {
						break // exit loop and close connection
					}
				}
			}

		} else if com == "name" { //------------------------------------------------------------------------------< NAME
			var payload = strings.TrimSpace(line[4:])
			payload = strings.ReplaceAll(payload, "\n", "") // remove protocol break
			payload = strings.ReplaceAll(payload, "\r", "") // remove protocol break
			if len(payload) >= 1 && len(payload) <= 25 {
				name = payload
				if comWrite(conn, "ok") {
					break // exit loop and close connection
				}
			} else {
				if comWrite(conn, "err: invalid name length") {
					break // exit loop and close connection
				}
			}

		} else if com == "type" { //----------------------------------------------------------------------< TYPE (color)
			var pl = strings.ToLower(strings.TrimSpace(line[4:]))
			if pl == "blue" || pl == "gray" || pl == "orange" || pl == "purple" || pl == "red" {
				color = pl
				if comWrite(conn, "ok") {
					break // exit loop and close connection
				}
			} else {
				if comWrite(conn, "err: invalid color; use 'blue', 'gray', 'orange', 'purple' or 'red'") {
					break // exit loop and close connection
				}
			}

		} else if com == "move" { //------------------------------------------------------------------------------< MOVE
			if me == nil {
				if comWrite(conn, "err: you're not playing") {
					break // exit loop and close connection
				}
			} else {
				if !ser.ready() {
					if comWrite(conn, "err: wait for other players") {
						break // exit loop and close connection
					}
				} else {
					// split input
					a := strings.Split(line[4:], ";")
					if len(a) != 2 {
						if comWrite(conn, "err: invalid input: use 'float32;float32'") {
							break // exit loop and close connection
						}
					}
					// parse value
					x, _ := strconv.ParseFloat(a[0], 32)
					y, _ := strconv.ParseFloat(a[1], 32)
					v := core.NewVelocity(float32(x), float32(y))
					// set wind
					if !ser.world.Move(me, v) {
						if comWrite(conn, "err: invalid move") {
							break // exit loop and close connection
						}
					} else {
						if comWrite(conn, "ok") {
							break // exit loop and close connection
						}
					}
				}
			}

		} else { // ---- default: invalid command -------------------------------------------------------------< DEFAULT
			if comWrite(conn, "err: invalid command") {
				break // exit loop and close connection
			}
		}
	}
}

func comWrite(conn net.Conn, s string) (error bool) {
	_, err := conn.Write([]byte(fmt.Sprintf("%s\r\n", s)))
	if err != nil {
		fmt.Printf("comWrite: %v\n", err)
		return true
	} else {
		return false
	}
}

func (ser *server) registerPlayer(name string) error {
	ser.mux.Lock()
	defer ser.mux.Unlock()

	// check names
	for _, n := range ser.players {
		if n == name {
			return errors.New("name already taken")
		}
	}

	// check maxPlayer
	if ser.ready() {
		return errors.New("maximum number of players reached")
	}

	// add player
	ser.players = append(ser.players, name)

	// un-freeze
	if ser.ready() {
		ser.world.Freeze(false)
	}

	// return success
	return nil
}

func (ser *server) ready() bool {
	return len(ser.players) >= ser.waitPlayer
}
