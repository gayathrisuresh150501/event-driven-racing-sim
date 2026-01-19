package engine

import (
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventbus"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventstore"
)

type Engine struct {
	state      *State
	eventBus   *eventbus.Bus
	eventStore *eventstore.Store
}

func New() *Engine {
	return &Engine{
		state:      NewState(),
		eventBus:   eventbus.New(),
		eventStore: eventstore.New(),
	}
}

// Tick advances the simulation by one step
func (e *Engine) Tick() {
	Advance(e.state, e.eventBus, e.eventStore)
}

// GetState returns the current engine state
// Used for inspection and testing
func (e *Engine) GetState() *State {
	return e.state
}

// EventBus returns the event bus for subscribing
func (e *Engine) EventBus() *eventbus.Bus {
	return e.eventBus
}

// EventStore returns the event store for replay
func (e *Engine) EventStore() *eventstore.Store {
	return e.eventStore
}