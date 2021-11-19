package simai

import (
	"CloudWars/core"
	"CloudWars/remote"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
)

// RunSimAI is a simple AI that accesses a game server with
// the remote.TcpClient and controls a player cloud.
// The AI calculates the future (8 sec) of 64 random
// movement commands every 100 ms and executes the best option.
func RunSimAI(host, port, name, color string) {

	// CONFIG ------------------------------------------
	var cpus = runtime.NumCPU()
	var angleSteps = 360 / 90
	var strengths = []float32{10, 50, 100, 200, 300}
	var simSpeedUp = 10
	var simInterval = 250 * time.Millisecond
	//--------------------------------------------------

	// prepare actions (list of move commands)
	actions := actionsSplitList(cpus, angleSteps, strengths)

	// connect to server and start game
	tcpClient := startGame(host, port, name, color)

	// ai loop
	for { //------------------------------------------------------------------------------------------------------------
		deadline := time.Now().Add(simInterval)

		// get new world status & calc future
		originWorld := loadStatus(tcpClient, simSpeedUp, simInterval)
		me := originWorld.Me(name)

		// start go simulations
		{ //---------------------------------
			wg := new(sync.WaitGroup)
			wg.Add(cpus) // Add a count of two, one for each goroutine.

			// go routines
			for i := 0; i < cpus; i++ {
				aa := actions[i]
				aa.ResetResults() // clear old results
				go simulate(i, wg, deadline, originWorld, name, aa)
			}

			// wait
			wg.Wait() // wait for simulations
			for !deadline.Before(time.Now()) {
				time.Sleep(100 * time.Microsecond) // wait for timeout
			}
		} //---------------------------------

		// find best result
		a := evaluation(actions, me)
		tcpClient.Move(a.Wind)

		// OPTIONAL: exit loop
		if me.IsDeath() {
			break
		}
	} //----------------------------------------------------------------------------------------------------------------

	// exit loop
	println("END", name)
	tcpClient.Kill()
	tcpClient.Close()
}

func startGame(host, port, name, color string) *remote.TcpClient {
	// init tcp client
	t := remote.NewTcpClient(host, port)

	// set name and start game
	t.Name(name)
	t.Color(color)
	resp := t.Play()

	// check resp
	if !strings.HasPrefix(resp, "ok") {
		log.Fatal(resp)
	}

	return t
}

func loadStatus(tcpClient *remote.TcpClient, simSpeedUp int, simInterval time.Duration) *core.World {
	// get new world
	json := tcpClient.List()
	w := &core.World{}
	w.FromJson(json)

	// simulate future
	ticks := simInterval.Seconds() * float64(w.GameSpeed())
	for i := float64(0); i <= ticks; i++ {
		w.Update()
	}

	// set sim speed
	w.SimSpeedUp = simSpeedUp

	// return
	return w
}
