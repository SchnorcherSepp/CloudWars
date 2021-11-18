package simai

import (
	"CloudWars/core"
	"math/rand"
)

type actions []*action

// ResetResults removes all results for the next simulation
func (aa actions) ResetResults() {
	for _, a := range aa {
		a.EvaluationPoints = 0
		a.ShortTerm = actionResult{}
		a.MidTerm = actionResult{}
		a.LongTerm = actionResult{}
	}
}

//--------------------------------------------------------------------------------------------------------------------//

type action struct {
	// action
	Wind     *core.Velocity
	Strength float32
	// results
	EvaluationPoints float64
	ShortTerm        actionResult
	MidTerm          actionResult
	LongTerm         actionResult
}

//--------------------------------------------------------------------------------------------------------------------//

type actionResult struct {
	StartIteration uint64
	EndIteration   uint64
	UsedWind       float32
	DeadEnemies    int
	GainVapor      float32
	GainSpeed      float32
	Alive          bool
}

//--------------------------------------------------------------------------------------------------------------------//

func actionsSplitList(cpus, angleSteps int, strengths []float32) []actions {

	// init basic list
	all := make(actions, 0)

	// add first action : NOTHING (wind = 0)
	all = append(all, &action{Wind: core.NewVelocity(0, 0), Strength: 0})

	// add random moves to basis list
	for _, strength := range strengths {
		aa := make(actions, 0)
		for i := 0; i < 360; i += angleSteps {
			a := &action{
				Wind:     core.NewVelocityByAngle(float32(i), strength),
				Strength: strength,
			}
			aa = append(aa, a)
		}
		rand.Shuffle(len(aa), func(i, j int) { aa[i], aa[j] = aa[j], aa[i] })
		all = append(all, aa...)
	}

	// prepare return list
	ret := make([]actions, cpus)
	for i := 0; i < cpus; i++ {
		ret[i] = make(actions, 0)
	}

	// split basic list
	i := 0
	for _, a := range all {
		ret[i] = append(ret[i], a)
		i++
		if i >= cpus {
			i = 0
		}
	}

	// return
	return ret
}
