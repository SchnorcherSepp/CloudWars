package gui

import (
	"CloudWars/core"
	"CloudWars/remote"
	"log"
	"os"
	"strings"
	"time"
)

// ModeClientGUI creates a GUI to participate in a remote game.
//
// server
//    host: server ip/host
//    port: server port
//
// Mode
//    observer: watching only
//
// player (gui)
//    name: name for player
//    color: color for player ('blue', 'gray', 'orange', 'purple' or 'red')
func ModeClientGUI(host, port string, observer bool, name, color string) {

	// init
	tcpClient := remote.NewTcpClient(host, port)
	cWorld := new(core.World)

	// CLIENT update loop
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			cWorld.FromJson(tcpClient.List())
		}
	}()

	// only with remote player
	if !observer {
		// start remote game
		if e := tcpClient.Name(name); !strings.HasPrefix(e, "ok") {
			log.Fatalf("ModeClientGUI: %v\n", e)
		}
		if e := tcpClient.Color(color); !strings.HasPrefix(e, "ok") {
			log.Fatalf("ModeClientGUI: %v\n", e)
		}
		if e := tcpClient.Play(); !strings.HasPrefix(e, "ok") {
			log.Fatalf("ModeClientGUI: %v\n", e)
		}
	}

	// need new cloud list from server with local player (me)
	// and for Width and Height
	cWorld.FromJson(tcpClient.List())

	// local player
	var me *core.Cloud
	if !observer {
		me = cWorld.Me(name)
	}

	// generate title
	title := "CloudWar Client  -  "
	if observer {
		title += "observer mode"
	} else {
		title += "remote play"
	}

	// CLIENT GUI
	if err := RunGame(title, cWorld.Width(), cWorld.Height(), cWorld.GameSpeed(), cWorld, me, false, tcpClient); err != nil {
		log.Fatalf("ModeClientGUI: %v\n", err)
	}

	// exit
	tcpClient.Kill()
	tcpClient.Close()
	os.Exit(0)
}
