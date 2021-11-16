package main

/*
This file contains an example in Go for an AI controlled client.
Use this example to program your own AI in Go.
*/

import (
	"CloudWars/core"
	"CloudWars/remote"
	"strings"
	"time"
)

// config
const (
	host  = "127.0.0.1"
	port  = "3333"
	name  = "BerndAI"
	color = "blue"
)

func main() {
	// init client
	tc := remote.NewTcpClient(host, port)

	// set name and start game
	tc.Name(name)
	tc.Color(color)
	if resp := tc.Play(); !strings.HasPrefix(resp, "ok") {
		println(resp)
		return
	}

	// get world status
	json := tc.List()
	world := &core.World{}
	world.FromJson(json)

	// get some world stats
	worldWidth := world.Width()         // game board size (DEFAULT: 2048)
	worldHeight := world.Height()       // game board size (DEFAULT: 1152)
	worldGameSpeed := world.GameSpeed() // updates per second (DEFAULT: 60)
	worldClouds := world.Clouds()       // cloud list

	// worldIteration: increases with every server update
	// worldVapor: vapor of all clouds together
	// worldAlive: active clouds
	worldIteration, worldVapor, worldAlive := world.Stats()

	// cloud list
	var me *core.Cloud // your controlled cloud (find in list)
	for _, cloud := range worldClouds {
		cloudName := cloud.Player // only player controlled clouds have names
		cloudColor := cloud.Color // cloud color
		cloudVapor := cloud.Vapor // cloud vapor (mass)
		cloudPosX := cloud.Pos.X  // x position
		cloudPosY := cloud.Pos.Y  // y position
		cloudVelX := cloud.Vel.X  // x velocity (speed)
		cloudVelY := cloud.Vel.Y  // y velocity (speed)
		if cloudName == name {
			me = cloud // set 'me'
		}

		// ignore this
		_ = cloudColor
		_ = cloudVapor
		_ = cloudPosX
		_ = cloudPosY
		_ = cloudVelX
		_ = cloudVelY
	}

	// make some decisions
	// move to the center
	if me.Pos.X < float32(worldWidth) {
		time.Sleep(2 * time.Second)
		tc.Move(core.NewVelocityByAngle(180, 33)) // move right
	} else {
		time.Sleep(2 * time.Second)
		tc.Move(core.NewVelocityByAngle(0, 33)) // move left
	}

	// move around
	time.Sleep(2 * time.Second)
	tc.Move(core.NewVelocityByAngle(0, 10)) // move left
	time.Sleep(2 * time.Second)
	tc.Move(core.NewVelocityByAngle(90, 10)) // move up
	time.Sleep(2 * time.Second)
	tc.Move(core.NewVelocityByAngle(180, 10)) // move right
	time.Sleep(2 * time.Second)
	tc.Move(core.NewVelocityByAngle(270, 10)) // move down
	time.Sleep(2 * time.Second)

	// it makes no sense
	tc.Kill()
	tc.Close()

	// ignore this
	_ = worldHeight
	_ = worldGameSpeed
	_ = worldIteration
	_ = worldVapor
	_ = worldAlive
}
