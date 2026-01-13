package replay

import (
	"testing"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

func TestReplaySimpleSequence(t *testing.T) {
	replayer := New()

	// Create a sequence of events
	eventList := []events.Event{
		events.NewTickAdvanced(1),
		events.NewCarMoved("car-1", 0, 0, 0, 1),
		events.NewTickAdvanced(2),
		events.NewCarMoved("car-1", 0, 1, 0, 2),
	}

	// Replay events
	state, err := replayer.Replay(eventList)
	if err != nil {
		t.Fatalf("replay failed: %v", err)
	}

	// Verify final state
	if state.GetTickNumber() != 2 {
		t.Errorf("expected tick 2, got %d", state.GetTickNumber())
	}

	car, ok := state.Cars["car-1"]
	if !ok {
		t.Fatal("car-1 not found")
	}

	if car.X != 0 || car.Y != 2 {
		t.Errorf("expected car at (0,2), got (%d,%d)", car.X, car.Y)
	}
}

func TestReplayEmptyEventList(t *testing.T) {
	replayer := New()

	state, err := replayer.Replay([]events.Event{})
	if err != nil {
		t.Fatalf("replay of empty list failed: %v", err)
	}

	// Should return initial state
	if state.GetTickNumber() != 0 {
		t.Errorf("expected tick 0 for empty replay, got %d", state.GetTickNumber())
	}
}

func TestReplayDetectsInconsistency(t *testing.T) {
	replayer := New()

	// Create events with wrong old position
	eventList := []events.Event{
		events.NewTickAdvanced(1),
		events.NewCarMoved("car-1", 0, 0, 0, 1),   // Correct
		events.NewCarMoved("car-1", 0, 999, 0, 2), // Wrong old position!
	}

	_, err := replayer.Replay(eventList)
	if err == nil {
		t.Error("expected replay to detect position inconsistency")
	}
}

func TestCompareStatesIdentical(t *testing.T) {
	// Create two engines and tick them identically
	e1 := engine.New()
	e2 := engine.New()

	e1.Tick()
	e2.Tick()

	// States should be identical
	if err := CompareStates(e1.GetState(), e2.GetState()); err != nil {
		t.Errorf("identical states reported as different: %v", err)
	}
}

func TestCompareStatesDifferent(t *testing.T) {
	e1 := engine.New()
	e2 := engine.New()

	// Tick e1 once, e2 twice
	e1.Tick()
	e2.Tick()
	e2.Tick()

	// States should be different
	if err := CompareStates(e1.GetState(), e2.GetState()); err == nil {
		t.Error("different states reported as identical")
	}
}
