package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	aipackage "github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// TestMultipleAIAgentsRaceAutonomously demonstrates multiple AI-controlled cars racing
func TestMultipleAIAgentsRaceAutonomously(t *testing.T) {
	e := engine.New()

	// Add multiple cars to the race
	e.GetState().Cars["car-2"] = &model.Car{ID: "car-2", X: 0, Y: 0, Speed: 1}
	e.GetState().Cars["car-3"] = &model.Car{ID: "car-3", X: 0, Y: 0, Speed: 1}
	e.GetState().Cars["car-4"] = &model.Car{ID: "car-4", X: 0, Y: 0, Speed: 1}

	// Register AI agents for each car with 100ms timeout protection
	e.RegisterAI("car-1", ai.NewTimeoutWrapper(ai.NewRuleBot("Alpha"), 100*time.Millisecond, 1))
	e.RegisterAI("car-2", ai.NewTimeoutWrapper(ai.NewRuleBot("Bravo"), 100*time.Millisecond, 1))
	e.RegisterAI("car-3", ai.NewTimeoutWrapper(ai.NewRuleBot("Charlie"), 100*time.Millisecond, 1))
	e.RegisterAI("car-4", ai.NewTimeoutWrapper(ai.NewRuleBot("Delta"), 100*time.Millisecond, 1))

	// Run the race for 20 ticks
	for i := 0; i < 20; i++ {
		e.Tick()
	}

	// Verify all cars moved forward
	for id, car := range e.GetState().Cars {
		if car.Y <= 0 {
			t.Errorf("car %s did not move forward: Y=%d", id, car.Y)
		}
		t.Logf("Car %s (AI: %s) final position: Y=%d, Speed=%d",
			id, e.GetAIAgents()[id].Name(), car.Y, car.Speed)
	}

	// Verify events were generated
	expectedMinEvents := 20 * (1 + 4) // 20 ticks * (1 TickAdvanced + 4 CarMoved per tick)
	if e.EventStore().Count() < expectedMinEvents {
		t.Errorf("expected at least %d events, got %d", expectedMinEvents, e.EventStore().Count())
	}

	t.Logf("✅ Multiple AI agents successfully raced autonomously for 20 ticks")
	t.Logf("   Generated %d events", e.EventStore().Count())
}

// TestEnginePerformanceWithSlowAI proves engine completes ticks quickly despite slow AI
// This is THE KEY TEST proving: AI suggests, engine decides. Engine NEVER waits for AI.
func TestEnginePerformanceWithSlowAI(t *testing.T) {
	e := engine.New()

	// Add cars
	e.GetState().Cars["car-2"] = &model.Car{ID: "car-2", X: 0, Y: 0, Speed: 1}

	// Create slow AI that takes 10 seconds (simulating expensive ML inference)
	slowAI := &SlowAI{delay: 10 * time.Second}

	// Wrap with 100ms timeout - this is critical for performance
	e.RegisterAI("car-1", ai.NewTimeoutWrapper(slowAI, 100*time.Millisecond, 1))
	e.RegisterAI("car-2", ai.NewTimeoutWrapper(ai.NewRuleBot("Fast"), 100*time.Millisecond, 1))

	// Measure time for 5 ticks
	start := time.Now()

	for i := 0; i < 5; i++ {
		e.Tick()
	}

	elapsed := time.Since(start)

	// With 5 ticks and 100ms timeout per tick, should complete in < 1 second
	// (allowing generous buffer for overhead)
	maxExpected := 1 * time.Second
	if elapsed > maxExpected {
		t.Errorf("Engine took too long with slow AI: %v (expected <%v)", elapsed, maxExpected)
	}

	// Verify both cars moved (slow AI falls back to default movement)
	car1 := e.GetState().Cars["car-1"]
	car2 := e.GetState().Cars["car-2"]

	if car1.Y <= 0 {
		t.Error("car-1 (slow AI) did not move")
	}

	if car2.Y <= 0 {
		t.Error("car-2 (fast AI) did not move")
	}

	t.Logf("✅ Engine completed 5 ticks in %v despite 10s slow AI (with 100ms timeout protection)", elapsed)
	t.Logf("   car-1 (slow AI with fallback): Y=%d", car1.Y)
	t.Logf("   car-2 (fast AI): Y=%d", car2.Y)
}

// SlowAI is a test AI that simulates expensive computation
type SlowAI struct {
	delay time.Duration
}

func (s *SlowAI) Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (*aipackage.Decision, error) {
	select {
	case <-time.After(s.delay):
		return &aipackage.Decision{
			CarID:         car.ID,
			SuggestedDeltaY: 100, // Would move very fast if not timed out
			Confidence:    1.0,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SlowAI) Name() string {
	return "SlowAI"
}

// TestAIDecisionsAffectMovement verifies AI decisions actually control car behavior
func TestAIDecisionsAffectMovement(t *testing.T) {
	e := engine.New()

	// Create a custom AI that suggests speed 3
	customAI := &FixedSpeedAI{speed: 3}
	e.RegisterAI("car-1", customAI)

	// Run 1 tick
	e.Tick()

	car := e.GetState().Cars["car-1"]

	// Car should have moved by 3 (AI's suggestion)
	expectedY := 3
	if car.Y != expectedY {
		t.Errorf("expected car Y=%d (AI suggested speed), got Y=%d", expectedY, car.Y)
	}

	if car.Speed != 3 {
		t.Errorf("expected car Speed=%d, got Speed=%d", 3, car.Speed)
	}

	t.Logf("✅ AI decisions successfully control car movement")
}

// FixedSpeedAI always suggests the same speed
type FixedSpeedAI struct {
	speed int
}

func (f *FixedSpeedAI) Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (*aipackage.Decision, error) {
	return &aipackage.Decision{
		CarID:         car.ID,
		SuggestedDeltaY: f.speed,
		Confidence:    1.0,
	}, nil
}

func (f *FixedSpeedAI) Name() string {
	return "FixedSpeedAI"
}
