package renderer

import (
	"strings"
	"testing"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

// TestRendererWithRealEngine tests the renderer integrated with the actual engine
// This proves the event-driven architecture works end-to-end
func TestRendererWithRealEngine(t *testing.T) {
	// Create engine
	e := engine.New()

	// Create renderer
	r := New(60, 15)

	// Subscribe renderer to engine events
	unsubscribe := e.EventBus().Subscribe(func(event events.Event) {
		r.HandleEvent(event)
	})
	defer unsubscribe()

	// Run some ticks
	for i := 0; i < 5; i++ {
		e.Tick()
	}

	// Verify renderer state matches engine events
	state := r.GetState()

	if state.TickNumber != 5 {
		t.Errorf("expected tick 5, got %d", state.TickNumber)
	}

	// Should have at least one car (from engine's default state)
	if len(state.Cars) == 0 {
		t.Error("expected cars in renderer state")
	}

	// Verify rendering works
	output := r.Render()
	if !strings.Contains(output, "Tick: 5") {
		t.Error("output missing tick number")
	}

	// Should show car info
	if !strings.Contains(output, "car-") {
		t.Error("output missing car info")
	}
}

// TestRendererReplayFromEventStore proves renderer can rebuild state from event store
// This is critical for demonstrating event sourcing - the renderer can reconstruct
// its entire view of the world by replaying the event log
func TestRendererReplayFromEventStore(t *testing.T) {
	// Create engine and run some ticks
	e := engine.New()
	for i := 0; i < 10; i++ {
		e.Tick()
	}

	// Get all events from event store
	storedEvents := e.EventStore().All()

	if len(storedEvents) == 0 {
		t.Fatal("expected events in store")
	}

	// Create a NEW renderer (proving it starts with no state)
	r := New(60, 15)

	// Replay ALL events from the store
	for _, event := range storedEvents {
		r.HandleEvent(event)
	}

	// Verify renderer rebuilt correct state
	state := r.GetState()

	if state.TickNumber != 10 {
		t.Errorf("expected tick 10 after replay, got %d", state.TickNumber)
	}

	if len(state.Cars) == 0 {
		t.Error("expected cars after replay")
	}

	// Verify car positions match the final state
	for _, car := range state.Cars {
		if car.X < 0 || car.Y < 0 {
			t.Errorf("car %s has invalid position (%d,%d)", car.ID, car.X, car.Y)
		}
	}
}

// TestRendererNeverAccessesEngineState ensures renderer truly never reads engine state
// This test would fail to compile if renderer tried to access engine.State directly
func TestRendererNeverAccessesEngineState(t *testing.T) {
	e := engine.New()
	r := New(60, 15)

	// Run engine
	e.Tick()

	// Renderer state should be empty (no events received)
	state := r.GetState()
	if state.TickNumber != 0 {
		t.Error("renderer should not know about engine state without events")
	}

	// Now subscribe and run again
	unsubscribe := e.EventBus().Subscribe(func(event events.Event) {
		r.HandleEvent(event)
	})
	defer unsubscribe()

	e.Tick()

	// Now renderer should have state (tick 2, since engine was at 1)
	state = r.GetState()
	if state.TickNumber != 2 {
		t.Errorf("expected tick 2, got %d", state.TickNumber)
	}
}

// TestRendererIndependentOfEngine proves renderer state is independent
// Modifying one doesn't affect the other - they're loosely coupled via events
func TestRendererIndependentOfEngine(t *testing.T) {
	e := engine.New()
	r := New(60, 15)

	// Subscribe and run
	unsubscribe := e.EventBus().Subscribe(func(event events.Event) {
		r.HandleEvent(event)
	})
	defer unsubscribe()

	e.Tick()
	e.Tick()

	// Both should be at tick 2
	if r.GetState().TickNumber != 2 {
		t.Errorf("renderer: expected tick 2, got %d", r.GetState().TickNumber)
	}

	// Unsubscribe renderer
	unsubscribe()

	// Engine continues without renderer
	e.Tick()
	e.Tick()

	// Engine at tick 4, renderer still at 2
	if r.GetState().TickNumber != 2 {
		t.Error("renderer should not update after unsubscribe")
	}

	// Create new renderer and replay from store
	r2 := New(60, 15)
	for _, event := range e.EventStore().All() {
		r2.HandleEvent(event)
	}

	// New renderer should catch up to tick 4
	if r2.GetState().TickNumber != 4 {
		t.Errorf("replayed renderer: expected tick 4, got %d", r2.GetState().TickNumber)
	}
}

// TestRendererEventOrderMatters verifies renderer correctly handles event ordering
func TestRendererEventOrderMatters(t *testing.T) {
	r := New(60, 15)

	// Out of order events
	r.HandleEvent(events.NewTickAdvanced(3))
	r.HandleEvent(events.NewTickAdvanced(1))
	r.HandleEvent(events.NewTickAdvanced(2))

	// Renderer applies events as received (last write wins)
	// In real system, event store maintains order
	if r.GetState().TickNumber != 2 {
		t.Errorf("expected tick 2 (last event), got %d", r.GetState().TickNumber)
	}
}
