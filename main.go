package main

import (
	"CloudWars/core"
	"CloudWars/gui"
	"CloudWars/remote"
	"CloudWars/simai"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const VERSION = "1.0"

const (
	descMode            = "Select Mode  ['singleplayer', 'server', 'client' or 'simai']"
	descHost            = "hostname or ip  [DEFAULT: localhost]"
	descPort            = "tcp port  [DEFAULT: 3333]"
	descScreenWidth     = "screen & game board width  [DEFAULT: 2048]"
	descScreenHeight    = "screen & game board height  [DEFAULT: 1152]"
	descGameSpeed       = "game speed in updates per second  [DEFAULT: 60]"
	descPlayerVapor     = "player vapor  [DEFAULT: 600]"
	descNeutralAmount   = "neutral cloud amount  [DEFAULT: 100]"
	descNeutralMaxSpeed = "neutral cloud max speed  [DEFAULT: 7]"
	descNeutralMaxVapor = "neutral cloud max vapor  [DEFAULT: 200]"
	descHeadless        = "run server without gui (headless)  [DEFAULT false]"
	descLocalPlayer     = "enable local player (false = observer)  [DEFAULT: true]"
	descLocalName       = "local player name"
	descLocalColor      = "local player color  ['blue', 'gray', 'orange', 'purple' or 'red']"
	descRemotePlayer    = "enable remote player  [DEFAULT: false]"
	descRemoteAmount    = "remote player amount  [DEFAULT: 3]"
)

func main() {

	// parse flags
	flagMode := flag.String("mode", "", descMode)
	flagHost := flag.String("host", "", descHost)
	flagPort := flag.String("port", "", descPort)
	flagScreenWidth := flag.String("width", "", descScreenWidth)
	flagScreenHeight := flag.String("height", "", descScreenHeight)
	flagGameSpeed := flag.String("speed", "", descGameSpeed)
	flagPlayerVapor := flag.String("pVapor", "", descPlayerVapor)
	flagNeutralAmount := flag.String("nAmount", "", descNeutralAmount)
	flagNeutralMaxSpeed := flag.String("nSpeed", "", descNeutralMaxSpeed)
	flagNeutralMaxVapor := flag.String("nVapor", "", descNeutralMaxVapor)
	flagHeadless := flag.String("headless", "", descHeadless)
	flagLocalPlayer := flag.String("lPlayer", "", descLocalPlayer)
	flagLocalName := flag.String("lName", "", descLocalName)
	flagLocalColor := flag.String("lColor", "", descLocalColor)
	flagRemotePlayer := flag.String("rPlayer", "", descRemotePlayer)
	flagRemoteAmount := flag.String("rAmount", "", descRemoteAmount)
	flag.Parse()

	// print defaults
	if len(os.Args) <= 1 {
		println("CloudWars", VERSION)
		println("---------------")
		flag.PrintDefaults()
		println()
	}

	// --- start interactive CLI --- //

	// mode
	mode := getString(flagMode, descMode, []string{"singleplayer", "server", "client", "simai"}, nil)
	switch mode {
	case "server":
		// server
		host := getString(flagHost, descHost, nil, nil) // allow empty ip in server mode!
		port := getString(flagPort, descPort, nil, []string{""})
		// game
		screenWidth := getInt(flagScreenWidth, descScreenWidth, nil, []string{""})
		screenHeight := getInt(flagScreenHeight, descScreenHeight, nil, []string{""})
		gameSpeed := getInt(flagGameSpeed, descGameSpeed, nil, []string{""})
		playerVapor := getInt(flagPlayerVapor, descPlayerVapor, nil, []string{""})
		neutralAmount := getInt(flagNeutralAmount, descNeutralAmount, nil, []string{""})
		neutralMaxSpeed := getInt(flagNeutralMaxSpeed, descNeutralMaxSpeed, nil, []string{""})
		neutralMaxVapor := getInt(flagNeutralMaxVapor, descNeutralMaxVapor, nil, []string{""})
		headless := getBool(flagHeadless, descHeadless, nil, []string{""})
		// local player
		var localPlayer bool
		var localName string
		var localColor string
		if !headless {
			localPlayer = getBool(flagLocalPlayer, descLocalPlayer, nil, []string{""})
			if localPlayer {
				localName = getString(flagLocalName, descLocalName, nil, []string{""})
				localColor = getString(flagLocalColor, descLocalColor, []string{"blue", "gray", "orange", "purple", "red"}, nil)
			}
		}
		// remote player
		var remotePlayer bool
		var remoteAmount int
		if !headless {
			remotePlayer = getBool(flagRemotePlayer, descRemotePlayer, nil, []string{""})
		} else {
			remotePlayer = true
		}
		if remotePlayer || headless {
			remoteAmount = getInt(flagRemoteAmount, descRemoteAmount, nil, []string{""})
		}

		// START SERVER
		if !headless {
			gui.ModeServerGUI(host, port, screenWidth, screenHeight, gameSpeed, float32(playerVapor), neutralAmount, float32(neutralMaxSpeed), float32(neutralMaxVapor), remotePlayer, remoteAmount, localPlayer, localName, localColor)
		} else {
			// create world
			sWorld := core.NewWorld(screenWidth, screenHeight, gameSpeed, neutralAmount, float32(neutralMaxSpeed), float32(neutralMaxVapor), time.Now().UnixMicro())
			// extern update loop
			go func() {
				for range time.Tick(1000 / time.Duration(gameSpeed) * time.Millisecond) {
					sWorld.Update()
				}
			}()
			// run server
			remote.RunServer(host, port, float32(playerVapor), sWorld, remoteAmount)
		}

	case "client":
		// server
		host := getString(flagHost, descHost, nil, []string{""})
		port := getString(flagPort, descPort, nil, []string{""})
		// local player
		localPlayer := getBool(flagLocalPlayer, descLocalPlayer, nil, []string{""})
		var localName string
		var localColor string
		if localPlayer {
			localName = getString(flagLocalName, descLocalName, nil, []string{""})
			localColor = getString(flagLocalColor, descLocalColor, []string{"blue", "gray", "orange", "purple", "red"}, nil)
		}

		// START CLIENT
		gui.ModeClientGUI(host, port, !localPlayer, localName, localColor)

	case "simai":
		// server
		host := getString(flagHost, descHost, nil, []string{""})
		port := getString(flagPort, descPort, nil, []string{""})
		// SimAI
		localColor := getString(flagLocalColor, descLocalColor, []string{"blue", "gray", "orange", "purple", "red"}, nil)
		localName := fmt.Sprintf("SimAI-%s", localColor)

		// START SimAI (client)
		simai.RunSimAI(host, port, localName, localColor)

	case "singleplayer":
		gui.ModeServerGUI("", "", 2048, 1152, 60, 600, 100, 7, 200, false, 0, true, "Cloudy", "blue")

	default:
		flag.PrintDefaults()
		log.Fatalf("err: main: invalid mode: %s", mode)
	}
}

//--------- HELPER ---------------------------------------------------------------------------------------------------//

func getString(flag *string, description string, whitelist, blacklist []string) string {
	var in = *flag

	// original input is ok
	if err := checkLists(in, whitelist, blacklist); err == "" {
		return in
	}

	// print description
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(description)
	fmt.Println("---------------------------------------")

	// read & check input
	for {
		// read
		fmt.Print("-> ")
		in, _ = reader.ReadString('\n')
		in = strings.ReplaceAll(in, "\n", "")
		in = strings.ReplaceAll(in, "\r", "")
		// check
		if err := checkLists(in, whitelist, blacklist); err != "" {
			fmt.Println(err) // nope
		} else {
			break // success
		}
	}

	// return new input
	return in
}

func getBool(flag *string, description string, whitelist, blacklist []string) bool {
	in := getString(flag, description, whitelist, blacklist)
	b, err := strconv.ParseBool(in)
	if err != nil {
		fmt.Printf("err: getBool: can't parse bool: %s\n", err)
	}
	return b
}

func getInt(flag *string, description string, whitelist, blacklist []string) int {
	in := getString(flag, description, whitelist, blacklist)
	i, err := strconv.ParseInt(in, 10, 32)
	if err != nil {
		fmt.Printf("err: getInt: can't parse int: %s\n", err)
	}
	return int(i)
}

func checkLists(in string, whitelist, blacklist []string) (err string) {
	// block invalid input
	if blacklist != nil {
		for _, s := range blacklist {
			if s == in {
				// is on blacklist
				return "Input is not allowed"
			}
		}
	}

	// enforce valid input
	if whitelist != nil {
		ok := false
		for _, s := range whitelist {
			if s == in {
				ok = true // is ok
				break
			}
		}
		if !ok {
			// not on whitelist
			return fmt.Sprintf("Input is not allowed. Choose one of: %v", whitelist)
		}
	}

	// no error
	return ""
}
