package gui

import (
	"CloudWars/core"
	"CloudWars/remote"
	"fmt"
	"log"
	"os"
	"time"
)

// ModeServerGUI creates a GUI and run a local server.
//
// server
//    host: server ip/host
//    port: server port
//
// game
//    screenWidth: game board size (DEFAULT: 2048)
//    screenHeight: game board size (DEFAULT: 1152)
//    gameSpeed: updates per second (DEFAULT: 60)
//    playerVapor: vapor for new player (DEFAULT: 600)
//
// neutral clouds
//    neutralAmount: number of neutral objects (DEFAULT: 100)
//    neutralMaxSpeed: random [0 to n] initial speed (DEFAULT: 7)
//    neutralMaxVapor: random [0-n] vapor for neutral objects (DEFAULT: 200)
//
// remote player (server)
//    remotePlayer: enable remote player
//    remoteAmount: wait for n remote player
//
// local player (gui)
//    localPlayer: enable local player (false: server mode only)
//    localName: name for local player
//    localColor: color for local player ('blue', 'gray', 'orange', 'purple' or 'red')
func ModeServerGUI(host, port string, screenWidth, screenHeight, gameSpeed int, playerVapor float32, neutralAmount int, neutralMaxSpeed, neutralMaxVapor float32, remotePlayer bool, remoteAmount int, localPlayer bool, localName, localColor string) {

	// init
	sWorld := core.NewWorld(screenWidth, screenHeight, gameSpeed, neutralAmount, neutralMaxSpeed, neutralMaxVapor, time.Now().UnixMicro())

	var lPlayer *core.Cloud
	if localPlayer {
		lPlayer = sWorld.AddPlayer(localName, localColor, nil, playerVapor)
	}

	// run server
	if remotePlayer {
		go remote.RunServer(host, port, playerVapor, sWorld, remoteAmount)
	}

	// generate title
	title := "CloudWar Server  -  "
	if localPlayer {
		title += fmt.Sprintf("player %s", localName)
	}
	if localPlayer && remotePlayer {
		title += "  &  "
	}
	if remotePlayer {
		title += fmt.Sprintf("%d remote player", remoteAmount)
	}

	// SERVER GUI
	if err := RunGame(title, screenWidth, screenHeight, gameSpeed, sWorld, lPlayer, false, nil); err != nil {
		log.Fatalf("ModeServerGUI: %v\n", err)
	}

	// exit
	os.Exit(0)
}
