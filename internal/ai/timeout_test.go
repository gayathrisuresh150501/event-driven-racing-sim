package ai_test

import (
	"context"
	"testing"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/ai"
	aipackage "github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// SlowAI is a test AI that deliberately takes too long
type SlowAI struct {
	delay time.Duration
}

func (s *SlowAI) Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (*aipackage.Decision, error) {
	// Simulate slow computation
	select {
	case <-time.After(s.delay):
		return &aipackage.Decision{
			CarID:         car.ID,
			SuggestedDeltaY: 10,
			Confidence:    1.0,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SlowAI) Name() string {
	return "SlowAI"
}

// TestSlowAI proves engine completes ticks in <200ms despite 10s AI delay
// This is THE KEY TEST proving: AI suggests, engine decides. Engine NEVER waits for AI.
func TestSlowAI(t *testing.T) {
	// Create an AI that takes 10 seconds to decide
	slowAI := &SlowAI{delay: 10 * time.Second}

	// Wrap it with 100ms timeout
	wrappedAI := ai.NewTimeoutWrapper(slowAI, 100*time.Millisecond, 1)

	// Create a test car
	car := &model.Car{
		ID:    "test-car",
		X:     0,
		Y:     0,
		Speed: 1,
	}

	// Measure how long the AI decision takes
	start := time.Now()

	decision, err := wrappedAI.Decide(context.Background(), car, []*model.Car{car}, 1)

	elapsed := time.Since(start)

	// Decision must complete in <200ms (giving some buffer for system overhead)
	if elapsed > 200*time.Millisecond {
		t.Errorf("AI decision took too long: %v (expected <200ms)", elapsed)
	}

	t.Logf("✅ AI decision completed in %v (with 10s slow AI wrapped in 100ms timeout)", elapsed)

	// We should get a fallback decision, not the slow AI's decision
	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	if decision == nil {
		t.Fatal("expected fallback decision, got nil")
	}

	// Fallback decision should have DeltaY = 1 (as configured)
	if decision.SuggestedDeltaY != 1 {
		t.Errorf("expected fallback DeltaY=1, got %d", decision.SuggestedDeltaY)
	}

	// Fallback should have zero confidence
	if decision.Confidence != 0.0 {
		t.Errorf("expected fallback confidence=0.0, got %f", decision.Confidence)
	}
}

// TestFastAI verifies that fast AIs work normally
func TestFastAI(t *testing.T) {
	// Create a fast AI (1ms delay)
	fastAI := &SlowAI{delay: 1 * time.Millisecond}

	// Wrap it with 100ms timeout
	wrappedAI := ai.NewTimeoutWrapper(fastAI, 100*time.Millisecond, 1)

	car := &model.Car{
		ID:    "test-car",
		X:     0,
		Y:     0,
		Speed: 1,
	}

	start := time.Now()
	decision, err := wrappedAI.Decide(context.Background(), car, []*model.Car{car}, 1)
	elapsed := time.Since(start)

	// Should complete quickly
	if elapsed > 50*time.Millisecond {
		t.Errorf("Fast AI took too long: %v", elapsed)
	}

	// Should not error
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Should get the actual AI's decision (DeltaY=10)
	if decision.SuggestedDeltaY != 10 {
		t.Errorf("expected AI DeltaY=10, got %d", decision.SuggestedDeltaY)
	}

	// Should have full confidence
	if decision.Confidence != 1.0 {
		t.Errorf("expected confidence=1.0, got %f", decision.Confidence)
	}

	t.Logf("✅ Fast AI completed successfully in %v", elapsed)
}

// TestRuleBot verifies the simple rule-based AI
func TestRuleBot(t *testing.T) {
	ruleBot := ai.NewRuleBot("test")

	car := &model.Car{
		ID:    "test-car",
		X:     0,
		Y:     0,
		Speed: 1,
	}

	// Test tick 0
	decision, err := ruleBot.Decide(context.Background(), car, []*model.Car{car}, 0)
	if err != nil {
		t.Fatalf("RuleBot failed: %v", err)
	}

	if decision.SuggestedDeltaY != 1 {
		t.Errorf("tick 0: expected DeltaY=1, got %d", decision.SuggestedDeltaY)
	}

	// Test tick 10 (should increase speed)
	decision, err = ruleBot.Decide(context.Background(), car, []*model.Car{car}, 10)
	if err != nil {
		t.Fatalf("RuleBot failed: %v", err)
	}

	if decision.SuggestedDeltaY != 2 {
		t.Errorf("tick 10: expected DeltaY=2, got %d", decision.SuggestedDeltaY)
	}

	// Test tick 50 (should cap at 5)
	decision, err = ruleBot.Decide(context.Background(), car, []*model.Car{car}, 50)
	if err != nil {
		t.Fatalf("RuleBot failed: %v", err)
	}

	if decision.SuggestedDeltaY != 5 {
		t.Errorf("tick 50: expected DeltaY=5 (capped), got %d", decision.SuggestedDeltaY)
	}

	t.Logf("✅ RuleBot behaves correctly across different tick numbers")
}

// TestRuleBotCatchUp verifies the catch-up logic
func TestRuleBotCatchUp(t *testing.T) {
	ruleBot := ai.NewRuleBot("test")

	// Car that's behind
	behindCar := &model.Car{
		ID:    "car-behind",
		X:     0,
		Y:     0,
		Speed: 1,
	}

	// Car that's ahead
	aheadCar := &model.Car{
		ID:    "car-ahead",
		X:     0,
		Y:     50,
		Speed: 1,
	}

	allCars := []*model.Car{behindCar, aheadCar}

	// At tick 0, behind car should try to catch up
	decision, err := ruleBot.Decide(context.Background(), behindCar, allCars, 0)
	if err != nil {
		t.Fatalf("RuleBot failed: %v", err)
	}

	// Base speed is 1, but with catch-up bonus it should be 2
	if decision.SuggestedDeltaY != 2 {
		t.Errorf("expected catch-up DeltaY=2, got %d", decision.SuggestedDeltaY)
	}

	t.Logf("✅ RuleBot correctly applies catch-up logic")
}
