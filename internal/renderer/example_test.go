package renderer

import (
	"fmt"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

// Example demonstrates how to use the renderer with events
func Example() {
	// Create renderer
	r := New(60, 15)

	// Simulate a race with events
	r.HandleEvent(events.NewTickAdvanced(1))
	r.HandleEvent(events.NewCarMoved("car-1", 10, 5, 11, 5))

	r.HandleEvent(events.NewTickAdvanced(2))
	r.HandleEvent(events.NewCarMoved("car-1", 11, 5, 12, 5))

	r.HandleEvent(events.NewTickAdvanced(3))
	r.HandleEvent(events.NewCarMoved("car-1", 12, 5, 13, 6))

	// Render current state
	output := r.Render()
	fmt.Print(output)

	// Note: Output would show the racing view with HUD
	// This example demonstrates the renderer works with only events
}

// ExampleRenderer_HandleEvent shows individual event handling
func ExampleRenderer_HandleEvent() {
	r := New(60, 15)

	// Process tick event
	r.HandleEvent(events.NewTickAdvanced(1))
	fmt.Printf("Tick: %d\n", r.GetState().TickNumber)

	// Process car movement event
	r.HandleEvent(events.NewCarMoved("car-1", 0, 0, 5, 3))
	car := r.GetState().Cars["car-1"]
	fmt.Printf("Car position: (%d, %d)\n", car.X, car.Y)
	fmt.Printf("Car speed: %d\n", car.Speed)

	// Output:
	// Tick: 1
	// Car position: (5, 3)
	// Car speed: 8
}

// ExampleState_ApplyEvent demonstrates state reconstruction from events
func ExampleState_ApplyEvent() {
	state := NewState()

	// Apply a sequence of events
	events := []events.Event{
		events.NewTickAdvanced(1),
		events.NewCarMoved("car-1", 0, 0, 1, 0),
		events.NewTickAdvanced(2),
		events.NewCarMoved("car-1", 1, 0, 2, 0),
		events.NewCarMoved("car-2", 0, 0, 0, 1),
	}

	for _, event := range events {
		state.ApplyEvent(event)
	}

	fmt.Printf("Tick: %d\n", state.TickNumber)
	fmt.Printf("Cars: %d\n", len(state.Cars))
	fmt.Printf("car-1: (%d, %d)\n", state.Cars["car-1"].X, state.Cars["car-1"].Y)
	fmt.Printf("car-2: (%d, %d)\n", state.Cars["car-2"].X, state.Cars["car-2"].Y)

	// Output:
	// Tick: 2
	// Cars: 2
	// car-1: (2, 0)
	// car-2: (0, 1)
}

// ExampleParseCommand shows command parsing
func ExampleParseCommand() {
	cmd := ParseCommand("tick")
	fmt.Printf("Command: %s, Is Tick: %t\n", cmd.Type, cmd.IsTick())

	cmd = ParseCommand("ticks 5")
	fmt.Printf("Command: %s, Args: %v\n", cmd.Type, cmd.Args)

	cmd = ParseCommand("quit")
	fmt.Printf("Command: %s, Is Quit: %t\n", cmd.Type, cmd.IsQuit())

	// Output:
	// Command: tick, Is Tick: true
	// Command: ticks, Args: [5]
	// Command: quit, Is Quit: true
}
