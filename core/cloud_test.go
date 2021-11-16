package core

import (
	"reflect"
	"testing"
)

func TestNewCloud(t *testing.T) {
	c1 := NewCloud(&World{height: 2}, &Position{X: 3}, &Velocity{X: 4}, 55, "Player 1", "Color 2")
	c2 := &Cloud{
		world:  &World{height: 2},
		Pos:    &Position{X: 3},
		Vel:    &Velocity{X: 4},
		Vapor:  55,
		Player: "Player 1",
		Color:  "Color 2",
	}

	// equal test
	if !reflect.DeepEqual(c1, c2) {
		t.Error("c1 and c2 not equal")
	}
}

func TestCloud_Radius(t *testing.T) {
	// radius of the cloud is always the square root of vapor
	if c := NewCloud(nil, nil, nil, 0*0, "", ""); c.Radius() != 0 {
		t.Errorf("fail: %v", c)
	}
	if c := NewCloud(nil, nil, nil, 1*1, "", ""); c.Radius() != 1 {
		t.Errorf("fail: %v", c)
	}
	if c := NewCloud(nil, nil, nil, 2*2, "", ""); c.Radius() != 2 {
		t.Errorf("fail: %v", c)
	}
	if c := NewCloud(nil, nil, nil, 1.5*1.5, "", ""); c.Radius() != 1.5 {
		t.Errorf("fail: %v", c)
	}
}

func TestCloud_IsDeath(t *testing.T) {
	if c := NewCloud(nil, nil, nil, 33, "", ""); c.IsDeath() { // alive
		t.Errorf("fail: %v", c)
	}
	if c := NewCloud(nil, nil, nil, -1, "", ""); !c.IsDeath() {
		t.Errorf("fail: %v", c)
	}
	if c := NewCloud(nil, nil, nil, 0, "", ""); !c.IsDeath() {
		t.Errorf("fail: %v", c)
	}
	if c := NewCloud(nil, nil, nil, 0.9999, "", ""); !c.IsDeath() {
		t.Errorf("fail: %v", c)
	}
	if c := NewCloud(nil, nil, nil, 1, "", ""); c.IsDeath() { // alive
		t.Errorf("fail: %v", c)
	}
}

func TestCloud_isIntersects(t *testing.T) {
	// vapor 100 is radius 10
	c1 := NewCloud(nil, &Position{X: 100, Y: 500}, nil, 100, "", "")
	c2 := NewCloud(nil, &Position{X: 200, Y: 500}, nil, 100, "", "")

	c2.Pos.X = 150 // fare away
	if c1.isIntersects(c2) {
		t.Errorf("fail:\n%v\t%f(%f)\n%v\t%f(%f)", c1.Pos, c1.Vapor, c1.Radius(), c2.Pos, c2.Vapor, c2.Radius())
	}

	c2.Pos.X = 120 // on the edge (c1 radius 10 AND c2 radio 10 IS 20)
	if c1.isIntersects(c2) {
		t.Errorf("fail:\n%v\t%f(%f)\n%v\t%f(%f)", c1.Pos, c1.Vapor, c1.Radius(), c2.Pos, c2.Vapor, c2.Radius())
	}

	c2.Pos.X = 119 // over
	if !c1.isIntersects(c2) {
		t.Errorf("fail:\n%v\t%f(%f)\n%v\t%f(%f)", c1.Pos, c1.Vapor, c1.Radius(), c2.Pos, c2.Vapor, c2.Radius())
	}

}

func TestCloud_clone(t *testing.T) {
	c1 := NewCloud(&World{width: 11}, &Position{X: 22}, &Velocity{X: 33}, 44, "player1", "color2")
	c2 := c1.clone()

	// Cloud.clone() don't clone Cloud.world !!!!!
	// fake it for DeepEqual
	c1.world = nil

	// equal test
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("vel1 and vel2 not equal:\n%v\n%v", c1, c2)
	}

	// changes to object 1 must not have any effects on object 2
	c1.Pos.Y = 99
	if reflect.DeepEqual(c1, c2) {
		t.Errorf("vel1 and vel2 not equal:\n%v\n%v", c1, c2)
	}
}

func TestCloud_update(t *testing.T) {
	w := NewWorld(100, 100, 60, 0, 0, 0, 0)
	c := NewCloud(w, NewPosition(20, 30), NewVelocity(10, 10), 100, "", "")
	c.update()

	// Movement (position += velocity * 0.1)
	if !almostEqual(c.Pos.X, 21) || !almostEqual(c.Pos.Y, 31) {
		t.Errorf("fail: %v", c.Pos)
	}

	// Damping of velocity (velocity *= 0.999)
	if c.Vel.X != 10*0.999 || c.Vel.Y != 10*0.999 {
		t.Errorf("fail: %v", c.Vel)
	}
}

func TestCloud_move(t *testing.T) {
	w := NewWorld(1000, 500, 60, 0, 0, 0, 0)
	c := NewCloud(w, NewPosition(500, 250), NewVelocity(10, 0), 100, "", "")
	w.addCloud(c)

	// This value is not allowed to be less than 1 or greater than vapor/2
	if c.move(NewVelocityByAngle(180, 0.99)) {
		t.Errorf("fail: %v", c)
	}
	if c.move(NewVelocityByAngle(180, 50.11)) {
		t.Errorf("fail: %v", c)
	}

	// move
	if !c.move(NewVelocityByAngle(0, 10)) {
		t.Errorf("fail: %v", c)
	}

	// The vapor property of the player controlled cloud will be reduced by Strength
	if c.Vapor != 100-10 {
		t.Errorf("fail: %v", c)
	}

	// The vector [(x / Radius) * 5, (y / Radius) * 5] is added to the velocity of the cloud
	//   wind X = 10
	//   radius = Sqrt(100-10)
	if !almostEqual(c.Vel.X, 10-(10*5/9.4868)) || !almostEqual(c.Vel.Y, 0) {
		t.Errorf("fail: %v", c.Vel)
	}

	// A new cloud is spawned with vapor equal to Strength
	if len(w.clouds) != 2 {
		t.Errorf("fail: %v", w.clouds)
	}
}

func TestCloud_kill(t *testing.T) {
	// init
	w := NewWorld(0, 0, 60, 0, 0, 0, 0)
	c := NewCloud(w, new(Position), new(Velocity), 33, "", "")
	if c.IsDeath() {
		t.Errorf("fail: %v", c)
	}

	// kill = true
	if !c.kill() {
		t.Errorf("fail: %v", c)
	}
	if c.Vapor != 0 {
		t.Errorf("fail: %v", c)
	}
	if !c.IsDeath() {
		t.Errorf("fail: %v", c)
	}

	// second kill = false
	c.Vapor = 0.99 // vapor < 1 is dead but not zero
	if c.kill() {
		t.Errorf("fail: %v", c)
	}
	if c.Vapor != 0 {
		t.Errorf("fail: %v", c)
	}
	if !c.IsDeath() {
		t.Errorf("fail: %v", c)
	}
}
