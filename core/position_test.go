package core

import (
	"reflect"
	"testing"
)

func TestNewPosition(t *testing.T) {
	pos1 := NewPosition(13, 9)
	pos2 := &Position{
		X: 13,
		Y: 9,
	}

	// equal test
	if !reflect.DeepEqual(pos1, pos2) {
		t.Error("pos1 and pos2 not equal")
	}
}

func TestPosition_clone(t *testing.T) {
	pos1 := NewPosition(69, 42)
	pos2 := pos1.clone()

	// equal test
	if !reflect.DeepEqual(pos1, pos2) {
		t.Error("pos1 and pos2 not equal")
	}

	// changes to object 1 must not have any effects on object 2
	pos2.X = 1
	if reflect.DeepEqual(pos1, pos2) {
		t.Error("pos1 and pos2 is equal")
	}
}

func TestPosition_add(t *testing.T) {

	// nil test
	pos := NewPosition(69, 42)
	pos.add(nil, 0) // invalid velocity -> no change
	if pos.X != 69 || pos.Y != 42 {
		t.Errorf("nil test fail: %v", pos)
	}

	// zero test
	pos = NewPosition(69, 42)
	pos.add(&Velocity{X: 10, Y: 10}, 0)
	if pos.X != 69 || pos.Y != 42 { // add zero:  add(69*0=0)
		t.Errorf("zero test fail: %v", pos)
	}

	// add test
	pos = NewPosition(69, 42)
	pos.add(&Velocity{X: 10, Y: 10}, 1.5) // add 15
	if pos.X != 69+15 || pos.Y != 42+15 {
		t.Errorf("add test fail: %v", pos)
	}

	// sub test
	pos = NewPosition(69, 42)
	pos.add(&Velocity{X: 10, Y: 10}, -1.5) // add -15
	if pos.X != 69-15 || pos.Y != 42-15 {
		t.Errorf("add test fail: %v", pos)
	}
}
