package simai

import (
	"CloudWars/core"
	"fmt"
	"math"
)

func evaluation(actionsList []actions, me *core.Cloud) *action {

	var best = &action{
		Wind:             &core.Velocity{},
		Strength:         0,
		EvaluationPoints: -10000000,
		ShortTerm:        actionResult{GainVapor: -10000000, GainSpeed: -10000000},
		MidTerm:          actionResult{GainVapor: -10000000, GainSpeed: -10000000},
		LongTerm:         actionResult{GainVapor: -10000000, GainSpeed: -10000000},
	}

	// preselection
	for _, aa := range actionsList {
		for _, a := range aa { //--------------------------

			// Calculates the percentage of vapor increase.
			// Depending on the period, the additional vapor is weighted less.
			// maxPerIncrease = |-30|5|30|61|204|
			sTPerIncrease := a.ShortTerm.GainVapor / me.Vapor * 100
			mTPerIncrease := a.MidTerm.GainVapor / me.Vapor * 100 * 0.75 // correction factor
			lTPerIncrease := a.LongTerm.GainVapor / me.Vapor * 100 * 0.3 // correction factor
			maxPerIncrease := math.Max(math.Max(float64(sTPerIncrease), float64(mTPerIncrease)), float64(lTPerIncrease))
			a.EvaluationPoints += maxPerIncrease

			// the percentage output is deducted from the points
			// perWindUsage = |0|2|9|24|100|
			perWindUsage := float64(a.Strength) / float64(me.Vapor) * 100
			a.EvaluationPoints -= perWindUsage

			// -1000 der eigene kurz oder mittelfristige tod
			// -500 der eigene langfristige tod
			if !a.ShortTerm.Alive || !a.MidTerm.Alive {
				a.EvaluationPoints -= 1000
			} else if !a.LongTerm.Alive {
				a.EvaluationPoints -= 200
			}

			// +500 ein vernichteter feind
			if a.ShortTerm.DeadEnemies > 0 {
				a.EvaluationPoints += 50
			} else if a.MidTerm.DeadEnemies > 0 {
				a.EvaluationPoints += 40
			} else if a.LongTerm.DeadEnemies > 0 {
				a.EvaluationPoints += 25
			}

			// find best
			if a.EvaluationPoints > best.EvaluationPoints {
				best = a
			}
		} //-----------------------------------------------
	}

	// log action
	if best.Strength > 0 {
		fmt.Printf("%.0f Wind  for  %.0f Points    ", best.Strength, best.EvaluationPoints)
		if best.EvaluationPoints < 0 {
			fmt.Printf("escape!")
		}
		if best.ShortTerm.DeadEnemies > 0 || best.MidTerm.DeadEnemies > 0 {
			fmt.Printf("KILL enemies!")
		} else if best.LongTerm.DeadEnemies > 0 {
			fmt.Printf("hunt enemies!")
		}
		fmt.Printf("\n")
	}

	// return
	return best
}
