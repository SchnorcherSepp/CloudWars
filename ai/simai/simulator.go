package simai

import (
	"CloudWars/core"
	"fmt"
	"log"
	"sync"
	"time"
)

func simulate(routine int, wg *sync.WaitGroup, deadline time.Time, originWorld *core.World, playerName string, actions actions) {
	defer wg.Done() // mark done when exiting the function

	// starting conditions to compare the results later
	originMe := originWorld.Me(playerName)
	var startIteration, _, _, _, _ = originWorld.Stats()
	var startVapor = originMe.Vapor
	var startSpeed = originMe.Vel.Strength()
	var startEnemies = countEnemies(originWorld, playerName)

	// simulation loop
	simIter := 0
	for _, a := range actions {
		simIter++

		// deadline -> LOOP EXIT
		if deadline.Before(time.Now()) {
			break
		}

		// prepare
		w := originWorld.Clone() // clone fresh world
		me := w.Me(playerName)   // find me in the clone world
		w.Move(me, a.Wind)       // set action

		// simulate terms
		for term, sec := range []float64{0.6, 1.0, 2.0} {

			// call updates for this term
			ticks := sec * float64(w.GameSpeed()) / float64(w.SimSpeedUp)
			for t := 0; t < int(ticks); t++ {
				w.Update()
			}
			var endIteration, _, _, _, _ = w.Stats()

			// calc results for this term
			switch term {
			case 0:
				a.ShortTerm.StartIteration = startIteration
				a.ShortTerm.EndIteration = endIteration
				a.ShortTerm.UsedWind = a.Strength
				a.ShortTerm.DeadEnemies = startEnemies - countEnemies(w, playerName)
				a.ShortTerm.GainVapor = me.Vapor - startVapor
				a.ShortTerm.GainSpeed = me.Vel.Strength() - startSpeed
				a.ShortTerm.Alive = !me.IsDeath()
			case 1:
				a.MidTerm.StartIteration = startIteration
				a.MidTerm.EndIteration = endIteration
				a.MidTerm.UsedWind = a.Strength
				a.MidTerm.DeadEnemies = startEnemies - countEnemies(w, playerName)
				a.MidTerm.GainVapor = me.Vapor - startVapor
				a.MidTerm.GainSpeed = me.Vel.Strength() - startSpeed
				a.MidTerm.Alive = !me.IsDeath()
			case 2:
				a.LongTerm.StartIteration = startIteration
				a.LongTerm.EndIteration = endIteration
				a.LongTerm.UsedWind = a.Strength
				a.LongTerm.DeadEnemies = startEnemies - countEnemies(w, playerName)
				a.LongTerm.GainVapor = me.Vapor - startVapor
				a.LongTerm.GainSpeed = me.Vel.Strength() - startSpeed
				a.LongTerm.Alive = !me.IsDeath()
			default:
				log.Fatalf("err: simulate: invalid term: %d", term)
			}
		}

	}

	// simulation alarm
	if routine == 0 {
		p := float64(simIter) / float64(len(actions)) * 100.0
		if p < 100 {
			fmt.Printf("ALERT: Not enough computing power!  %.0f %%\n", p)
		}
	}
}

func countEnemies(world *core.World, playerName string) (enemies int) {
	for _, c := range world.Clouds() {
		if c.Player != "" && c.Player != playerName && !c.IsDeath() {
			enemies++
		}
	}
	return
}
