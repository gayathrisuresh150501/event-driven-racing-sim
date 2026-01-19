package renderer

import "github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"

// State represents the renderer's view of the world, built entirely from events
// CRITICAL: This state NEVER reads from engine.State directly
type State struct {
	TickNumber int
	Cars       map[string]*CarState
}

// CarState tracks a car's state as observed through events
type CarState struct {
	ID    string
	X     int
	Y     int
	Speed int // Speed is calculated from distance traveled per tick
}

// NewState creates a new renderer state
func NewState() *State {
	return &State{
		Cars: make(map[string]*CarState),
	}
}

// ApplyEvent updates the renderer state based on an event
// This is the ONLY way the renderer learns about the world
func (s *State) ApplyEvent(event events.Event) {
	switch e := event.(type) {
	case *events.TickAdvanced:
		s.TickNumber = e.TickNumber

	case *events.CarMoved:
		car, exists := s.Cars[e.CarID]
		if !exists {
			// First time seeing this car - create it
			car = &CarState{
				ID: e.CarID,
				X:  e.OldX,
				Y:  e.OldY,
			}
			s.Cars[e.CarID] = car
		}

		// Calculate speed from distance traveled
		dx := e.NewX - e.OldX
		dy := e.NewY - e.OldY
		distance := abs(dx) + abs(dy) // Manhattan distance
		car.Speed = distance

		// Update position
		car.X = e.NewX
		car.Y = e.NewY
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
