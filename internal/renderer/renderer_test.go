package renderer

import (
	"strings"
	"testing"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

// TestRendererReplay proves the renderer works with ONLY events
// It replays a sequence of events and verifies the renderer builds correct state
// CRITICAL: This test ensures the renderer NEVER needs direct engine state access
func TestRendererReplay(t *testing.T) {
	r := New(60, 15)

	// Simulate a sequence of events (could be from event store)
	eventSequence := []events.Event{
		events.NewTickAdvanced(1),
		events.NewCarMoved("car-1", 0, 0, 1, 0),
		events.NewTickAdvanced(2),
		events.NewCarMoved("car-1", 1, 0, 2, 0),
		events.NewTickAdvanced(3),
		events.NewCarMoved("car-1", 2, 0, 3, 0),
	}

	// Replay events through renderer
	for _, event := range eventSequence {
		r.HandleEvent(event)
	}

	// Verify renderer state matches what events describe
	state := r.GetState()

	if state.TickNumber != 3 {
		t.Errorf("expected tick 3, got %d", state.TickNumber)
	}

	if len(state.Cars) != 1 {
		t.Fatalf("expected 1 car, got %d", len(state.Cars))
	}

	car := state.Cars["car-1"]
	if car == nil {
		t.Fatal("car-1 not found")
	}

	if car.X != 3 || car.Y != 0 {
		t.Errorf("expected car at (3,0), got (%d,%d)", car.X, car.Y)
	}

	if car.Speed != 1 {
		t.Errorf("expected speed 1, got %d", car.Speed)
	}

	// Verify rendering works
	output := r.Render()
	if output == "" {
		t.Error("render produced empty output")
	}

	// Should contain tick number
	if !strings.Contains(output, "Tick: 3") {
		t.Error("render output missing tick number")
	}

	// Should contain car info
	if !strings.Contains(output, "car-1") {
		t.Error("render output missing car-1")
	}
}

// TestRendererMultipleCars tests rendering multiple cars from events
func TestRendererMultipleCars(t *testing.T) {
	r := New(60, 15)

	// Multiple cars moving
	events := []events.Event{
		events.NewTickAdvanced(1),
		events.NewCarMoved("car-1", 0, 0, 1, 0),
		events.NewCarMoved("car-2", 0, 0, 0, 1),
		events.NewTickAdvanced(2),
		events.NewCarMoved("car-1", 1, 0, 2, 0),
		events.NewCarMoved("car-2", 0, 1, 0, 2),
	}

	for _, event := range events {
		r.HandleEvent(event)
	}

	state := r.GetState()

	if len(state.Cars) != 2 {
		t.Fatalf("expected 2 cars, got %d", len(state.Cars))
	}

	car1 := state.Cars["car-1"]
	if car1.X != 2 || car1.Y != 0 {
		t.Errorf("car-1: expected (2,0), got (%d,%d)", car1.X, car1.Y)
	}

	car2 := state.Cars["car-2"]
	if car2.X != 0 || car2.Y != 2 {
		t.Errorf("car-2: expected (0,2), got (%d,%d)", car2.X, car2.Y)
	}

	// Verify both cars appear in output
	output := r.Render()
	if !strings.Contains(output, "car-1") || !strings.Contains(output, "car-2") {
		t.Error("render output missing cars")
	}
}

// TestRendererEventOrder tests that event order matters
func TestRendererEventOrder(t *testing.T) {
	r := New(60, 15)

	// Car moves, then tick advances
	r.HandleEvent(events.NewCarMoved("car-1", 0, 0, 5, 5))
	r.HandleEvent(events.NewTickAdvanced(1))

	state := r.GetState()

	// Tick should be 1 even though move came first
	if state.TickNumber != 1 {
		t.Errorf("expected tick 1, got %d", state.TickNumber)
	}

	// Car should be at (5,5)
	car := state.Cars["car-1"]
	if car.X != 5 || car.Y != 5 {
		t.Errorf("expected car at (5,5), got (%d,%d)", car.X, car.Y)
	}
}

// TestRendererSpeedCalculation tests speed is calculated from events
func TestRendererSpeedCalculation(t *testing.T) {
	r := New(60, 15)

	// Car moves different distances
	r.HandleEvent(events.NewTickAdvanced(1))
	r.HandleEvent(events.NewCarMoved("car-1", 0, 0, 2, 0)) // Distance: 2

	state := r.GetState()
	if state.Cars["car-1"].Speed != 2 {
		t.Errorf("expected speed 2, got %d", state.Cars["car-1"].Speed)
	}

	// Next tick, moves diagonal
	r.HandleEvent(events.NewTickAdvanced(2))
	r.HandleEvent(events.NewCarMoved("car-1", 2, 0, 4, 2)) // Distance: 2+2=4

	if state.Cars["car-1"].Speed != 4 {
		t.Errorf("expected speed 4, got %d", state.Cars["car-1"].Speed)
	}
}

// TestRendererRenderingOutput tests the visual output structure
func TestRendererRenderingOutput(t *testing.T) {
	r := New(60, 15)

	r.HandleEvent(events.NewTickAdvanced(1))
	r.HandleEvent(events.NewCarMoved("car-1", 10, 5, 10, 5))

	output := r.Render()

	// Should have borders
	if !strings.Contains(output, "╔") || !strings.Contains(output, "╚") {
		t.Error("output missing borders")
	}

	// Should have track markers
	if !strings.Contains(output, "│") {
		t.Error("output missing track markers")
	}

	// Should have HUD
	if !strings.Contains(output, "Tick:") {
		t.Error("output missing HUD")
	}

	// Should have car details
	if !strings.Contains(output, "Pos(") {
		t.Error("output missing car position info")
	}
}

// TestRendererEmptyState tests rendering with no events
func TestRendererEmptyState(t *testing.T) {
	r := New(60, 15)

	// No events processed yet
	state := r.GetState()

	if state.TickNumber != 0 {
		t.Errorf("expected tick 0, got %d", state.TickNumber)
	}

	if len(state.Cars) != 0 {
		t.Errorf("expected 0 cars, got %d", len(state.Cars))
	}

	// Should still render without crashing
	output := r.Render()
	if output == "" {
		t.Error("render produced empty output for empty state")
	}

	if !strings.Contains(output, "Cars: 0") {
		t.Error("output should show 0 cars")
	}
}
