package engine

import (
	"context"
	"sort"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventbus"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/eventstore"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// queryAIAgents asks all AI agents for their decisions
// This function respects the principle: AI suggests, engine decides. Engine NEVER waits for AI.
func queryAIAgents(state *State, aiAgents map[string]ai.RaceAI) map[string]*ai.Decision {
	decisions := make(map[string]*ai.Decision)

	if len(aiAgents) == 0 {
		return decisions
	}

	// Prepare context with tight timeout (should be enforced by wrapper, but belt-and-suspenders)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	// Build list of all cars for AI context
	allCars := make([]*model.Car, 0, len(state.Cars))
	for _, car := range state.Cars {
		allCars = append(allCars, car)
	}

	// Query each AI agent
	for carID, agent := range aiAgents {
		car, exists := state.Cars[carID]
		if !exists {
			continue // Car doesn't exist, skip
		}

		// Ask AI for decision (non-blocking with timeout)
		decision, _ := agent.Decide(ctx, car, allCars, state.TickNumber)

		// Store decision (even if nil, we track that we asked)
		decisions[carID] = decision
	}

	return decisions
}

// applyAIDecision applies an AI's suggested decision to a car
func applyAIDecision(car *model.Car, decision *ai.Decision) {
	if decision == nil {
		return
	}

	// AI suggests Y-axis movement
	car.Y += decision.SuggestedDeltaY

	// Update the car's speed to match the actual movement
	car.Speed = decision.SuggestedDeltaY
}

func Advance(state *State, bus *eventbus.Bus, store *eventstore.Store, aiAgents map[string]ai.RaceAI) {
	// Increment tick
	state.TickNumber++

	// Emit TickAdvanced event
	tickEvent := events.NewTickAdvanced(state.TickNumber)
	if bus != nil {
		bus.Publish(tickEvent)
	}
	if store != nil {
		if err := store.Append(tickEvent); err != nil {
			panic(err) // Should never happen with valid events
		}
	}

	// Query AI agents for decisions (non-blocking)
	// AI suggests, engine decides. Engine NEVER waits for AI.
	aiDecisions := queryAIAgents(state, aiAgents)

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

		// Apply AI decision if available
		if decision, ok := aiDecisions[id]; ok && decision != nil {
			applyAIDecision(car, decision)
		} else {
			// Fallback to default movement if no AI decision
			applyMovement(car)
		}

		// Emit CarMoved event using the map key as canonical ID
		moveEvent := events.NewCarMoved(id, oldX, oldY, car.X, car.Y)
		if bus != nil {
			bus.Publish(moveEvent)
		}
		if store != nil {
			if err := store.Append(moveEvent); err != nil {
				panic(err) // Should never happen with valid events
			}
		}
	}
}
