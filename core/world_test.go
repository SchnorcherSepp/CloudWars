package core

import (
	"reflect"
	"testing"
)

func initTestWorld() *World {
	w := NewWorld(6666, 3333, 60, 150, 50, 800, 1337)
	w.AddPlayer("Player 1", "red", NewPosition(300, 300), 1000)
	w.AddPlayer("Player 2", "blue", NewPosition(1300, 1300), 1000)
	return w
}

func TestWorld_Serialisation(t *testing.T) {

	// init world
	origin := initTestWorld()
	json := origin.ToJson()

	// clone world
	clone := new(World)
	clone.FromJson(json)

	// check state 0
	if !reflect.DeepEqual(origin, clone) {
		t.Error("DeepEqual fail in state 0")
	}

	// update origin and check (origin state 1; clone state 0)
	origin.Update()
	if reflect.DeepEqual(origin, clone) {
		t.Error("is DeepEqual in state 1 & 0")
	}

	// update clone (state 1)
	clone.Update()
	if !reflect.DeepEqual(origin, clone) {
		t.Error("DeepEqual fail in state 1")
	}

	// check second "FromJson" (state 501)
	for i := 0; i < 500; i++ {
		origin.Update()
	}
	clone.FromJson(origin.ToJson())
	if !reflect.DeepEqual(origin, clone) {
		t.Error("DeepEqual fail in state 501")
	}
}

func TestNewWorld(t *testing.T) {
	w1 := NewWorld(111, 222, 60, 5, 0, 200, 1337)
	w1.Update()

	w2 := &World{
		width:        111,
		height:       222,
		iteration:    1,
		worldVapor:   703.94507, // random vapor (see seed 1337)
		alive:        5,
		winCondition: true,
		leader:       "no player alive",
		clouds:       nil, // fix this
		mux:          nil, // fix this
		SimSpeedUp:   1,
		gameSpeed:    60,
	}

	// fix clout und mux
	w1.clouds = nil
	w1.mux = nil

	// equal test
	if !reflect.DeepEqual(w1, w2) {
		t.Error("w1 and w2 not equal")
	}
}

func TestWorld_Clone(t *testing.T) {
	w1 := NewWorld(2222, 1111, 60, 50, 30, 600, 1337)
	w1.Update()
	w2 := w1.Clone()
	w1.Update()
	w3 := w2.Clone()

	// equal test
	if reflect.DeepEqual(w1, w2) {
		t.Errorf("w1 and w2 equal:\n%v\n%v", w1, w2)
	}

	// equal test
	if reflect.DeepEqual(w1, w3) {
		t.Errorf("w1 and w3 equal:\n%v\n%v", w1, w3)
	}

	// equal test
	if !reflect.DeepEqual(w2, w3) {
		t.Errorf("w2 and w3 not equal:\n%v\n%v", w2, w3)
	}

}

func TestWorld_Clouds(t *testing.T) {
	// TODO: implement
	//       clone!
}

func TestWorld_Me(t *testing.T) {
	// TODO: implement
	//       no clone, direct ref!
}

func TestWorld_AddPlayer(t *testing.T) {
	// TODO: implement
}

func TestWorld_Move(t *testing.T) {
	// TODO: implement
}

func TestWorld_Kill(t *testing.T) {
	// TODO: implement
}

func TestWorld_Update(t *testing.T) {
	// TODO: implement
}
