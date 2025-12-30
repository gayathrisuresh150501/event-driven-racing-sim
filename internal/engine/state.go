package engine

import "github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"

type State struct {
	Cars map[string]*model.Car
}

func NewState() *State {
	return &State{
		Cars: map[string]*model.Car{
			"car-1": {ID: "car-1", X: 0, Y: 0, Speed: 1},
		},
	}
}
