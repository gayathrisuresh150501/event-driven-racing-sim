package ai

import (
	"context"
	"fmt"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// RuleBot is a simple rule-based AI that makes decisions based on basic logic
type RuleBot struct {
	name string
}

// NewRuleBot creates a new rule-based AI agent
func NewRuleBot(name string) *RuleBot {
	return &RuleBot{
		name: name,
	}
}

// Decide implements the RaceAI interface
// Strategy: Move forward at a speed that increases every 10 ticks, capped at 5
func (r *RuleBot) Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (*ai.Decision, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Calculate desired speed based on tick number
	// Start at speed 1, increase by 1 every 10 ticks, max speed 5
	desiredSpeed := 1 + (tickNumber / 10)
	if desiredSpeed > 5 {
		desiredSpeed = 5
	}

	// If we're behind, try to catch up by going faster
	// Check if any car is significantly ahead
	maxY := car.Y
	for _, otherCar := range allCars {
		if otherCar.ID != car.ID && otherCar.Y > maxY {
			maxY = otherCar.Y
		}
	}

	// If we're more than 10 units behind the leader, boost speed
	if maxY-car.Y > 10 {
		desiredSpeed++
		if desiredSpeed > 5 {
			desiredSpeed = 5
		}
	}

	// Return the decision
	return &ai.Decision{
		CarID:         car.ID,
		SuggestedDeltaY: desiredSpeed,
		Confidence:    0.8, // Reasonably confident in this simple strategy
	}, nil
}

// Name returns the AI's identifier
func (r *RuleBot) Name() string {
	return fmt.Sprintf("RuleBot-%s", r.name)
}
