package core

import (
	"math"
	"reflect"
	"testing"
)

func almostEqual(a, b float32) bool {
	const float64EqualityThreshold = 0.0001
	return math.Abs(float64(a)-float64(b)) <= float64EqualityThreshold
}

func TestNewVelocity(t *testing.T) {
	vel1 := NewVelocity(13, 9)
	vel2 := &Velocity{
		X: 13,
		Y: 9,
	}

	// equal test
	if !reflect.DeepEqual(vel1, vel2) {
		t.Error("vel1 and vel2 not equal")
	}
}

func TestNewVelocityByAngle(t *testing.T) {
	// 0° -> move left
	if vel := NewVelocityByAngle(0, 1); !almostEqual(vel.X, -1) || !almostEqual(vel.Y, 0) {
		t.Errorf("angle fail: %v", vel)
	}

	// 90° -> move up
	if vel := NewVelocityByAngle(90, 1); !almostEqual(vel.X, 0) || !almostEqual(vel.Y, -1) {
		t.Errorf("angle fail: %v", vel)
	}

	// 180° -> move right
	if vel := NewVelocityByAngle(180, 1); !almostEqual(vel.X, 1) || !almostEqual(vel.Y, 0) {
		t.Errorf("angle fail: %v", vel)
	}

	// 270° -> move down
	if vel := NewVelocityByAngle(270, 1); !almostEqual(vel.X, 0) || !almostEqual(vel.Y, 1) {
		t.Errorf("angle fail: %v", vel)
	}

	// 360°  ==  0
	if vel := NewVelocityByAngle(360, 1); !almostEqual(vel.X, -1) || !almostEqual(vel.Y, 0) {
		t.Errorf("angle fail: %v", vel)
	}

	//  45° -> move left/up (Pythagoras)
	if vel := NewVelocityByAngle(45, 1.41421); !almostEqual(vel.X, -1) || !almostEqual(vel.Y, -1) {
		t.Errorf("angle fail: %v", vel)
	}
}

func TestVelocity_Strength(t *testing.T) {
	// Pythagoras

	if vel := NewVelocity(0, 0); !almostEqual(vel.Strength()*vel.Strength(), 0) {
		t.Errorf("pythagoras fail: %f", vel.Strength()*vel.Strength())
	}

	if vel := NewVelocity(1, 1); !almostEqual(vel.Strength()*vel.Strength(), 1*1+1*1) {
		t.Errorf("pythagoras fail: %f", vel.Strength()*vel.Strength())
	}

	if vel := NewVelocity(2, 0); !almostEqual(vel.Strength()*vel.Strength(), 2*2+0*0) {
		t.Errorf("pythagoras fail: %f", vel.Strength()*vel.Strength())
	}

	if vel := NewVelocity(-5, 3); !almostEqual(vel.Strength()*vel.Strength(), -5*(-5)+3*3) {
		t.Errorf("pythagoras fail: %f", vel.Strength()*vel.Strength())
	}
}

func TestVelocity_clone(t *testing.T) {
	vel1 := NewVelocity(69, 42)
	vel2 := vel1.clone()

	// equal test
	if !reflect.DeepEqual(vel1, vel2) {
		t.Error("vel1 and vel2 not equal")
	}

	// changes to object 1 must not have any effects on object 2
	vel2.X = 1
	if reflect.DeepEqual(vel1, vel2) {
		t.Error("vel1 and vel2 is equal")
	}
}

func TestVelocity_add(t *testing.T) {

	// nil test
	vel := NewVelocity(69, 42)
	vel.add(nil, 0) // invalid velocity -> no change
	if vel.X != 69 || vel.Y != 42 {
		t.Errorf("nil test fail: %v", vel)
	}

	// zero test
	vel = NewVelocity(69, 42)
	vel.add(&Velocity{X: 10, Y: 10}, 0)
	if vel.X != 69 || vel.Y != 42 { // add zero:  add(69*0=0)
		t.Errorf("zero test fail: %v", vel)
	}

	// add test
	vel = NewVelocity(69, 42)
	vel.add(&Velocity{X: 10, Y: 10}, 1.5) // add 15
	if vel.X != 69+15 || vel.Y != 42+15 {
		t.Errorf("add test fail: %v", vel)
	}

	// sub test
	vel = NewVelocity(69, 42)
	vel.add(&Velocity{X: 10, Y: 10}, -1.5) // add -15
	if vel.X != 69-15 || vel.Y != 42-15 {
		t.Errorf("sub test fail: %v", vel)
	}
}

func TestVelocity_multi(t *testing.T) {

	vel := NewVelocity(69, 42)
	vel.multi(0)
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("multi fail: %v", vel)
	}

	vel = NewVelocity(69, 42)
	vel.multi(1)
	if vel.X != 69 || vel.Y != 42 {
		t.Errorf("multi fail: %v", vel)
	}

	vel = NewVelocity(69, 42)
	vel.multi(-1.5)
	if vel.X != -69*1.5 || vel.Y != -42*1.5 {
		t.Errorf("multi fail: %v", vel)
	}
}
