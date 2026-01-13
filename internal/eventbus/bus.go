package eventbus

import (
	"sync"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

// Handler is a function that handles events
type Handler func(events.Event)

// Bus is an in-memory event bus
type Bus struct {
	handlers map[int]Handler
	nextID   int
	mu       sync.RWMutex
}

// New creates a new event bus
func New() *Bus {
	return &Bus{
		handlers: make(map[int]Handler),
		nextID:   1,
	}
}

// Subscribe adds a handler to receive all events and returns an unsubscribe function
func (b *Bus) Subscribe(handler Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.handlers[id] = handler
	b.nextID++

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.handlers, id)
	}
}

// Publish sends an event to all subscribers
func (b *Bus) Publish(event events.Event) {
	b.mu.RLock()
	// Copy handlers to slice to avoid holding lock during execution
	handlers := make([]Handler, 0, len(b.handlers))
	for _, h := range b.handlers {
		handlers = append(handlers, h)
	}
	b.mu.RUnlock()

	// Call handlers synchronously to maintain determinism
	for _, handler := range handlers {
		handler(event)
	}
}

// Clear removes all handlers (useful for testing)
func (b *Bus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = make(map[int]Handler)
}
