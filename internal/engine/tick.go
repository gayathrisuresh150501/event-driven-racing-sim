package engine

import (
	"sort"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventbus"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventstore"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

func Advance(state *State, bus *eventbus.Bus, store *eventstore.Store) {
	// Increment tick
	state.TickNumber++
	
	// Emit TickAdvanced event
	tickEvent := events.NewTickAdvanced(state.TickNumber)
	bus.Publish(tickEvent)
	store.Append(tickEvent)
	
	// Move each car and emit events
	// Sort car IDs to ensure deterministic event generation order
	carIDs := make([]string, 0, len(state.Cars))
	for id := range state.Cars {
		carIDs = append(carIDs, id)
	}
	sort.Strings(carIDs)

	for _, id := range carIDs {
		car := state.Cars[id]
		// Record old position
		oldX := car.X
		oldY := car.Y
		
		// Apply movement
		applyMovement(car)
		
		// Emit CarMoved event
		moveEvent := events.NewCarMoved(car.ID, oldX, oldY, car.X, car.Y)
		bus.Publish(moveEvent)
		store.Append(moveEvent)
	}
}
