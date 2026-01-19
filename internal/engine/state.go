package engine

import "github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"

type State struct {
	Cars       map[string]*model.Car
	TickNumber int
}

func NewState() *State {
	return &State{
		Cars: map[string]*model.Car{
			"car-1": {ID: "car-1", X: 0, Y: 0, Speed: 1},
		},
		TickNumber: 0,
	}
}

// GetTickNumber returns the current tick number
func (s *State) GetTickNumber() int {
	return s.TickNumber
}

// SetTickNumber sets the tick number (used by replay)
func (s *State) SetTickNumber(tick int) {
	s.TickNumber = tick
}