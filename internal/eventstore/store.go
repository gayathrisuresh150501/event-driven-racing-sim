package eventstore

import (
	"sync"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

// Store is an append-only event log
type Store struct {
	events []events.Event
	mu     sync.RWMutex
}

// New creates a new event store
func New() *Store {
	return &Store{
		events: make([]events.Event, 0),
	}
}

// Append adds an event to the store
// Events are immutable once added
func (s *Store) Append(event events.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

// All returns all events in order
// Returns a copy to prevent external mutation
func (s *Store) All() []events.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	copied := make([]events.Event, len(s.events))
	copy(copied, s.events)
	return copied
}

// Count returns the number of events
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

// Since returns all events starting at the given index (inclusive).
// Returns an empty slice if index is out of bounds (negative or >= event count).
func (s *Store) Since(index int) []events.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index < 0 || index >= len(s.events) {
		return []events.Event{}
	}

	copied := make([]events.Event, len(s.events)-index)
	copy(copied, s.events[index:])
	return copied
}

// Clear removes all events (useful for testing)
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = make([]events.Event, 0)
}
