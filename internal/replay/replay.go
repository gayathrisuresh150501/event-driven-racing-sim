package replay

import (
	"fmt"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// Replayer rebuilds state from events
type Replayer struct{}

// New creates a new replayer
func New() *Replayer {
	return &Replayer{}
}

// Replay rebuilds engine state from a sequence of events
// This proves that events are the source of truth
func (r *Replayer) Replay(eventList []events.Event) (*engine.State, error) {
	// Start with initial state
	state := engine.NewState()
	
	// Apply each event in order
	for i, event := range eventList {
		if err := r.applyEvent(state, event); err != nil {
			return nil, fmt.Errorf("failed to apply event %d (%s): %w", i, event.EventType(), err)
		}
	}
	
	return state, nil
}

// applyEvent applies a single event to the state
func (r *Replayer) applyEvent(state *engine.State, event events.Event) error {
	switch e := event.(type) {
	case *events.TickAdvanced:
		// Update tick number
		state.SetTickNumber(e.TickNumber)
		return nil
		
	case *events.CarMoved:
		// Apply car movement
		car, ok := state.Cars[e.CarID]
		if !ok {
			return fmt.Errorf("car %s not found", e.CarID)
		}
		
		// Verify old position matches (optional consistency check)
		if car.X != e.OldX || car.Y != e.OldY {
			return fmt.Errorf("car %s position mismatch: expected (%d,%d), got (%d,%d)",
				e.CarID, e.OldX, e.OldY, car.X, car.Y)
		}
		
		// Apply new position
		car.X = e.NewX
		car.Y = e.NewY
		return nil
		
	default:
		// Unknown event type - this is fine, just skip it
		// This allows for forward compatibility
		return nil
	}
}

// CompareStates compares two states for equality
// Used in tests to verify replay produces the same result
func CompareStates(s1, s2 *engine.State) error {
	if s1.GetTickNumber() != s2.GetTickNumber() {
		return fmt.Errorf("tick number mismatch: %d vs %d", s1.GetTickNumber(), s2.GetTickNumber())
	}
	
	if len(s1.Cars) != len(s2.Cars) {
		return fmt.Errorf("car count mismatch: %d vs %d", len(s1.Cars), len(s2.Cars))
	}
	
	for carID, car1 := range s1.Cars {
		car2, ok := s2.Cars[carID]
		if !ok {
			return fmt.Errorf("car %s missing in second state", carID)
		}
		
		if err := compareCars(car1, car2); err != nil {
			return fmt.Errorf("car %s mismatch: %w", carID, err)
		}
	}
	
	return nil
}

func compareCars(c1, c2 *model.Car) error {
	if c1.ID != c2.ID {
		return fmt.Errorf("ID mismatch: %s vs %s", c1.ID, c2.ID)
	}
	if c1.X != c2.X {
		return fmt.Errorf("X mismatch: %d vs %d", c1.X, c2.X)
	}
	if c1.Y != c2.Y {
		return fmt.Errorf("Y mismatch: %d vs %d", c1.Y, c2.Y)
	}
	if c1.Speed != c2.Speed {
		return fmt.Errorf("Speed mismatch: %d vs %d", c1.Speed, c2.Speed)
	}
	return nil
}