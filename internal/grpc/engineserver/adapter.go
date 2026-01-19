package engineserver

import (
	"context"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// remoteAIAdapter adapts a remote AIClient to the local ai.RaceAI interface
type remoteAIAdapter struct {
	client AIClient
}

// Decide implements the ai.RaceAI interface
func (r *remoteAIAdapter) Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (*ai.Decision, error) {
	// Add timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
	}

	decision, err := r.client.Decide(ctx, car, allCars, tickNumber)
	if err != nil {
		return nil, err
	}

	return &ai.Decision{
		CarID:           decision.CarID,
		SuggestedDeltaY: decision.SuggestedDeltaY,
		Confidence:      decision.Confidence,
	}, nil
}

// Name implements the ai.RaceAI interface
func (r *remoteAIAdapter) Name() string {
	return r.client.Name()
}
