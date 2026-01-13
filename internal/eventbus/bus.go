package eventbus

import (
	"sync"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

// Handler is a function that handles events
type Handler func(events.Event)

// Bus is an in-memory event bus
type Bus struct {
	handlers []Handler
	mu       sync.RWMutex
}

// New creates a new event bus
func New() *Bus {
	return &Bus{
		handlers: make([]Handler, 0),
	}
}

// Subscribe adds a handler to receive all events
func (b *Bus) Subscribe(handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

// Publish sends an event to all subscribers
func (b *Bus) Publish(event events.Event) {
	b.mu.RLock()
	handlers := b.handlers
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
	b.handlers = make([]Handler, 0)
}