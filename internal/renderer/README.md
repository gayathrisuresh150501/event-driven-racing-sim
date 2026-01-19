# Renderer Package

Event-driven terminal renderer for the racing simulator.

## Architecture

This package implements a pure event-driven renderer that **NEVER** reads engine state directly. All state is built from events.

For complete architecture details, design patterns, usage examples, and testing documentation, see:

**[RENDERER_ARCHITECTURE.md](../../RENDERER_ARCHITECTURE.md)** in the project root.

## Quick Start

```go
engine := engine.New()
renderer := renderer.New(60, 15)

// Subscribe renderer to events
unsubscribe := engine.EventBus().Subscribe(func(event events.Event) {
    renderer.HandleEvent(event)
})
defer unsubscribe()

// Run simulation
engine.Tick()
fmt.Println(renderer.Render())
```

## Package Files

- `state.go` - Event-sourced state management
- `renderer.go` - Terminal rendering logic
- `input.go` - Command input handling
- `renderer_test.go` - Unit tests
- `integration_test.go` - Integration tests
- `example_test.go` - Example code
