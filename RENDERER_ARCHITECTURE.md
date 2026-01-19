# Event-Driven Renderer Architecture

This document describes the architecture of the terminal renderer for the racing simulator, demonstrating pure event-driven design.

## Core Principle: Event-Only State

**The renderer NEVER reads engine state directly.** Instead, it maintains its own view of the world built entirely from events.

### Why This Matters

This architecture demonstrates several critical concepts:

1. **Loose Coupling**: Renderer and engine are completely independent
2. **Event Sourcing**: State can be rebuilt from event log
3. **Time Travel**: Can replay events to any point in history
4. **Multiple Views**: Multiple renderers can show same data differently
5. **Testability**: Can test renderer with synthetic events (no engine needed)

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Engine                              │
│  - Manages simulation state                                 │
│  - Emits events on state changes                            │
│  - NEVER directly coupled to renderer                       │
└────────────┬────────────────────────────────────────────────┘
             │
             │ Events (TickAdvanced, CarMoved, etc.)
             ▼
┌─────────────────────────────────────────────────────────────┐
│                       Event Bus                             │
│  - Distributes events to subscribers                        │
│  - Synchronous delivery (for determinism)                   │
│  - Manages subscriptions/unsubscriptions                    │
└────┬───────────────────────────┬──────────────────────┬─────┘
     │                           │                      │
     │ Events                    │ Events               │ Events
     ▼                           ▼                      ▼
┌─────────────┐        ┌──────────────────┐    ┌──────────────┐
│ Event Store │        │     Renderer     │    │Other Listeners│
│- Persists   │        │- Subscribes      │    │              │
│  events     │        │  to events       │    │              │
│- Enables    │        │- Builds own      │    │              │
│  replay     │        │  state           │    │              │
└─────────────┘        │- Never reads     │    └──────────────┘
                       │  engine state    │
                       └──────────────────┘
```

## Component Details

### Renderer State (`internal/renderer/state.go`)

```go
type State struct {
    TickNumber int
    Cars       map[string]*CarState
}

type CarState struct {
    ID    string
    X     int
    Y     int
    Speed int  // Calculated from event deltas
}
```

**Key Methods:**
- `ApplyEvent(event)`: ONLY way to update state
- Never has direct access to `engine.State`

### Renderer (`internal/renderer/renderer.go`)

**Responsibilities:**
- Subscribe to event bus
- Update internal state via events
- Render visual representation
- Handle keyboard input (converted to commands)

**Does NOT:**
- Read from `engine.State`
- Directly modify engine
- Know about engine internals

### Input Handler (`internal/renderer/input.go`)

Converts keyboard input to **commands** (not direct mutations):

```go
type Command struct {
    Type string  // "tick", "quit", "help", etc.
    Args []string
}
```

Commands are requests to the engine, not direct state changes.

## Data Flow

### Forward Flow (Real-time)

```
1. User types "tick" command
2. Engine executes Tick()
3. Engine emits events (TickAdvanced, CarMoved)
4. Event Bus publishes events
5. Renderer receives events via subscription
6. Renderer updates its state
7. Renderer re-renders view
8. User sees updated display
```

### Replay Flow (Event Sourcing)

```
1. Load events from Event Store
2. Create fresh renderer (empty state)
3. Replay events through renderer
4. Renderer rebuilds state from scratch
5. Final state matches historical reality
```

## Critical Tests

### `TestRendererReplay` (renderer_test.go)

Proves renderer works with events only:

```go
func TestRendererReplay(t *testing.T) {
    r := New(60, 15)

    // Feed synthetic events (no engine!)
    events := []events.Event{
        events.NewTickAdvanced(1),
        events.NewCarMoved("car-1", 0, 0, 1, 0),
        // ...
    }

    for _, e := range events {
        r.HandleEvent(e)
    }

    // Verify state matches events
    assert(r.GetState().TickNumber == expected)
}
```

### `TestRendererReplayFromEventStore` (integration_test.go)

Proves event sourcing capability:

```go
func TestRendererReplayFromEventStore(t *testing.T) {
    // Run engine to generate events
    engine := engine.New()
    for i := 0; i < 10; i++ {
        engine.Tick()
    }

    // Create NEW renderer (empty state)
    renderer := New(60, 15)

    // Replay ALL events from store
    for _, event := range engine.EventStore().All() {
        renderer.HandleEvent(event)
    }

    // State should match final engine state
    // But renderer NEVER read engine state directly!
}
```

### `TestRendererNeverAccessesEngineState` (integration_test.go)

Ensures architecture purity:

```go
func TestRendererNeverAccessesEngineState(t *testing.T) {
    engine := engine.New()
    renderer := New(60, 15)

    // Engine runs, renderer not subscribed
    engine.Tick()

    // Renderer state is empty (no events received)
    assert(renderer.GetState().TickNumber == 0)

    // This proves renderer can't "cheat" by reading engine
}
```

## Visual Output

```
╔══════════════════════════════════════════════════════════╗
║ Tick: 10 | Cars: 1                                      ║
║ car-1: Pos(10,5) Speed:1                                ║
╠══════════════════════════════════════════════════════════╣
║          ░░░░░░░░░░░░│              │░░░░░░░░░░░░       ║
║         ░░░░░░░░░░░░│                │░░░░░░░░░░░░      ║
║        ░░░░░░░░░░░│                  │░░░░░░░░░░░       ║
║       ░░░░░░░░░░│                    │░░░░░░░░░░        ║
║      ░░░░░░░░░│          █           │░░░░░░░░░         ║
║     ░░░░░░░░│                        │░░░░░░░░          ║
║    ░░░░░░░│                          │░░░░░░░           ║
║   ░░░░░░│                            │░░░░░░            ║
║  ░░░░░│                              │░░░░░             ║
║ ░░░░│                                │░░░░              ║
╚══════════════════════════════════════════════════════════╝
```

**Features:**
- **HUD**: Tick number, car count, positions, speeds
- **Pseudo-3D Track**: Perspective view with narrowing track
- **Car Markers**: `█` symbols show car positions
- **Track Elements**: Borders (`│`), grass (`░`), center line

## Usage Examples

### Basic Usage

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

// Display
fmt.Println(renderer.Render())
```

### Event Replay

```go
// Rebuild state from event log
renderer := renderer.New(60, 15)
for _, event := range eventStore.All() {
    renderer.HandleEvent(event)
}
fmt.Println(renderer.Render())
```

### Multiple Renderers

```go
// Different views of same events
fullView := renderer.New(80, 20)
compactView := renderer.New(40, 10)

engine.EventBus().Subscribe(func(e events.Event) {
    fullView.HandleEvent(e)
    compactView.HandleEvent(e)
})
```

## Benefits Demonstrated

### 1. Loose Coupling

Renderer and engine can evolve independently. Changes to engine internals don't affect renderer as long as events remain consistent.

### 2. Event Sourcing

Complete state can be rebuilt from event log. No snapshots needed - events are the source of truth.

### 3. Time Travel

Can replay events to any point:
```go
// Show state at tick 100
for _, event := range store.Since(100) {
    renderer.HandleEvent(event)
}
```

### 4. Debugging

Replay problematic event sequences with logging:
```go
for _, event := range problematicEvents {
    log.Printf("Before: %+v", renderer.GetState())
    renderer.HandleEvent(event)
    log.Printf("After: %+v", renderer.GetState())
}
```

### 5. Testing

Test with synthetic events - no engine required:
```go
renderer.HandleEvent(events.NewTickAdvanced(999))
renderer.HandleEvent(events.NewCarMoved("test-car", 0, 0, 100, 100))
// Verify rendering logic without running engine
```

## Design Patterns

### Event Sourcing
- State derived from events
- Events are immutable
- Can replay to rebuild state

### CQRS (Command Query Responsibility Segregation)
- Commands (input) → Engine
- Events (output) → Renderer
- Separate read/write models

### Observer Pattern
- Engine is subject
- Renderer is observer
- Event bus is mediator

### Publish-Subscribe
- Engine publishes events
- Renderer subscribes to events
- Loose coupling via event bus

## Files

- `internal/renderer/state.go` - Event-sourced state
- `internal/renderer/renderer.go` - Main renderer logic
- `internal/renderer/input.go` - Command input handling
- `internal/renderer/renderer_test.go` - Unit tests
- `internal/renderer/integration_test.go` - Integration tests
- `internal/renderer/README.md` - Package documentation
- `cmd/renderer-demo/main.go` - Interactive demo

## Running the Demo

```bash
go run cmd/renderer-demo/main.go
```

Commands:
- `tick` - Advance by one tick
- `ticks N` - Advance by N ticks
- `help` - Show commands
- `quit` - Exit

## Key Takeaways

1. **Pure Event-Driven**: Renderer learns ONLY from events
2. **No Shortcuts**: Never reads engine state directly
3. **Event Sourcing**: Can rebuild from event log
4. **Testable**: Works with synthetic events
5. **Loosely Coupled**: Engine and renderer independent

This architecture proves complex UIs can run entirely on event streams without direct state access.
