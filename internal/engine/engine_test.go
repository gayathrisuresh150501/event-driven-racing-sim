package engine_test

import (
	"testing"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/replay"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

const ticksForDeterminismTest = 20

func TestDeterministicEngine(t *testing.T) {
	e1 := engine.New()
	e2 := engine.New()

	// Assert identical initial state
	c1, ok1 := e1.GetState().Cars["car-1"]
	c2, ok2 := e2.GetState().Cars["car-1"]

	if !ok1 || !ok2 {
		t.Fatalf("expected car-1 to exist in initial engine state")
	}

	if c1.X != c2.X || c1.Y != c2.Y || c1.Speed != c2.Speed {
		t.Fatalf("initial engine state differs: %+v vs %+v", c1, c2)
	}

	for i := 0; i < ticksForDeterminismTest; i++ {
		e1.Tick()
		e2.Tick()
	}

	f1, ok1 := e1.GetState().Cars["car-1"]
	f2, ok2 := e2.GetState().Cars["car-1"]

	if !ok1 || !ok2 {
		t.Fatalf("expected car-1 to exist after ticks")
	}

	if f1.X != f2.X || f1.Y != f2.Y || f1.Speed != f2.Speed {
		t.Fatalf(
			"engine state is not deterministic after ticks: %+v vs %+v",
			f1,
			f2,
		)
	}
}

// TestEventEmission verifies events are emitted during tick
func TestEventEmission(t *testing.T) {
	e := engine.New()

	eventCount := 0
	e.EventBus().Subscribe(func(event events.Event) {
		eventCount++
	})

	e.Tick()

	// Should emit: 1 TickAdvanced + 1 CarMoved
	expectedEvents := 2
	if eventCount != expectedEvents {
		t.Errorf("expected %d events, got %d", expectedEvents, eventCount)
	}
}

// TestEventStore verifies events are stored
func TestEventStore(t *testing.T) {
	e := engine.New()

	e.Tick()
	e.Tick()

	// Should have: 2 TickAdvanced + 2 CarMoved = 4 events
	expectedEvents := 4
	if e.EventStore().Count() != expectedEvents {
		t.Errorf("expected %d stored events, got %d", expectedEvents, e.EventStore().Count())
	}
}

// TestReplayDeterminism - THE KEY TEST
// This proves events are the source of truth
func TestReplayDeterminism(t *testing.T) {
	// Create engine and run it
	e := engine.New()

	for i := 0; i < ticksForDeterminismTest; i++ {
		e.Tick()
	}

	// Get final state
	originalState := e.GetState()

	// Replay from events
	replayer := replay.New()
	replayedState, err := replayer.Replay(e.EventStore().All())
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	// States must be identical
	if err := replay.CompareStates(originalState, replayedState); err != nil {
		t.Errorf("replay did not produce identical state: %v", err)
	}

	t.Logf("✅ Replay succeeded: original and replayed states are identical after %d ticks", ticksForDeterminismTest)
}

// TestMultipleReplay verifies replay is deterministic
func TestMultipleReplay(t *testing.T) {
	// Create engine and run it
	e := engine.New()

	for i := 0; i < 10; i++ {
		e.Tick()
	}

	events := e.EventStore().All()

	// Replay multiple times
	replayer := replay.New()

	state1, err1 := replayer.Replay(events)
	if err1 != nil {
		t.Fatalf("first replay failed: %v", err1)
	}

	state2, err2 := replayer.Replay(events)
	if err2 != nil {
		t.Fatalf("second replay failed: %v", err2)
	}

	// Both replays must produce identical results
	if err := replay.CompareStates(state1, state2); err != nil {
		t.Errorf("replays produced different results: %v", err)
	}
}
