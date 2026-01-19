package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// TimeoutWrapper wraps any RaceAI and enforces a strict timeout
// If the wrapped AI exceeds the timeout, a fallback decision is returned
type TimeoutWrapper struct {
	wrapped        ai.RaceAI
	timeout        time.Duration
	fallbackDeltaY int // Default movement if AI times out or fails
}

// NewTimeoutWrapper creates a new AI wrapper with timeout protection
// timeout: maximum time allowed for AI to decide (recommended: 100ms)
// fallbackDeltaY: movement to use if AI fails (recommended: 1 for conservative forward movement)
func NewTimeoutWrapper(wrapped ai.RaceAI, timeout time.Duration, fallbackDeltaY int) *TimeoutWrapper {
	return &TimeoutWrapper{
		wrapped:        wrapped,
		timeout:        timeout,
		fallbackDeltaY: fallbackDeltaY,
	}
}

// Decide implements the RaceAI interface with timeout enforcement
func (t *TimeoutWrapper) Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (*ai.Decision, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	// Channel to receive the decision
	type result struct {
		decision *ai.Decision
		err      error
	}
	resultChan := make(chan result, 1)

	// Run the wrapped AI in a goroutine
	go func() {
		// Recover from any panics in the AI
		defer func() {
			if r := recover(); r != nil {
				resultChan <- result{
					decision: nil,
					err:      fmt.Errorf("AI panic: %v", r),
				}
			}
		}()

		decision, err := t.wrapped.Decide(timeoutCtx, car, allCars, tickNumber)
		resultChan <- result{decision: decision, err: err}
	}()

	// Wait for either the result or timeout
	select {
	case res := <-resultChan:
		// AI completed within timeout
		if res.err != nil {
			// AI returned an error, use fallback
			return t.fallbackDecision(car), fmt.Errorf("AI error (using fallback): %w", res.err)
		}
		return res.decision, nil

	case <-timeoutCtx.Done():
		// Timeout exceeded, return fallback decision
		return t.fallbackDecision(car), fmt.Errorf("AI timeout after %v (using fallback)", t.timeout)
	}
}

// fallbackDecision creates a safe default decision when AI fails
func (t *TimeoutWrapper) fallbackDecision(car *model.Car) *ai.Decision {
	return &ai.Decision{
		CarID:         car.ID,
		SuggestedDeltaY: t.fallbackDeltaY,
		Confidence:    0.0, // Zero confidence indicates fallback was used
	}
}

// Name returns the wrapped AI's name with timeout indicator
func (t *TimeoutWrapper) Name() string {
	return fmt.Sprintf("%s[timeout:%v]", t.wrapped.Name(), t.timeout)
}
