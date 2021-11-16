package remote

import (
	"CloudWars/core"
	"strings"
	"testing"
	"time"
)

func TestNewTcpClient(t *testing.T) {

	// init
	world := core.NewWorld(2000, 1000, 60, 30, 20, 400, 1337)
	world.Update()
	go RunServer("localhost", "8686", 800, world, 1)
	time.Sleep(1 * time.Second)
	client := NewTcpClient("localhost", "8686")

	// fail commands
	if res := client.Kill(); res != "err: you're not playing" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Move(&core.Velocity{}); res != "err: you're not playing" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Name(""); res != "err: invalid name length" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Color(""); res != "err: invalid color; use 'blue', 'gray', 'orange', 'purple' or 'red'" {
		t.Errorf("fail: %s", res)
	}

	// success
	if res := client.List(); !strings.HasPrefix(res, "{") {
		t.Errorf("fail: %s", res)
	}
	if res := client.Name("Hanspeter"); res != "ok" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Color("blue"); res != "ok" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Play(); res != "ok: the game begins when all players are ready" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Move(nil); res != "err: nil" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Move(core.NewVelocityByAngle(45, 3333)); res != "err: invalid move" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Move(core.NewVelocityByAngle(45, 33)); res != "ok" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Kill(); res != "ok" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Kill(); res != "err: you're already dead" {
		t.Errorf("fail: %s", res)
	}
	if res := client.Close(); res != "ok" {
		t.Errorf("fail: %s", res)
	}
}
