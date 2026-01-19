package engine

import (
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventbus"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventstore"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/ai"
)

type Engine struct {
	state      *State
	eventBus   *eventbus.Bus
	eventStore *eventstore.Store
	aiAgents   map[string]ai.RaceAI // Maps car ID to AI agent
}

func New() *Engine {
	return &Engine{
		state:      NewState(),
		eventBus:   eventbus.New(),
		eventStore: eventstore.New(),
		aiAgents:   make(map[string]ai.RaceAI),
	}
}

// RegisterAI registers an AI agent to control a specific car
func (e *Engine) RegisterAI(carID string, agent ai.RaceAI) {
	e.aiAgents[carID] = agent
}

// GetAIAgents returns the map of AI agents
func (e *Engine) GetAIAgents() map[string]ai.RaceAI {
	return e.aiAgents
}

// Tick advances the simulation by one step
func (e *Engine) Tick() {
	Advance(e.state, e.eventBus, e.eventStore, e.aiAgents)
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