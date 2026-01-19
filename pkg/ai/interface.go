package ai

import (
	"context"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// Decision represents an AI's suggested action for a car
type Decision struct {
	CarID         string
	SuggestedDeltaY int  // Suggested Y-axis movement
	Confidence    float64 // 0.0 to 1.0, how confident the AI is
}

// RaceAI is the contract for AI agents that control racing cars
// Key principle: AI suggests, engine decides. Engine NEVER waits for AI.
type RaceAI interface {
	// Decide is called by the engine to get the AI's suggested action
	// The AI receives:
	// - ctx: context for cancellation/timeout control
	// - car: current state of the car the AI is controlling
	// - allCars: complete view of all cars in the race
	// - tickNumber: current simulation tick
	//
	// The AI returns:
	// - Decision: the suggested action (or nil if no decision)
	// - error: any error encountered (timeout, panic recovery, etc.)
	//
	// CRITICAL: This function MUST respect context cancellation.
	// The engine will enforce timeouts to prevent slow AIs from blocking simulation.
	Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (*Decision, error)

	// Name returns a human-readable identifier for this AI
	Name() string
}
