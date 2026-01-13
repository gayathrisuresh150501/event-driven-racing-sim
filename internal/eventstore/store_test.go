package eventstore

import (
	"testing"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

func TestStoreAppend(t *testing.T) {
	store := New()
	
	event1 := events.NewTickAdvanced(1)
	event2 := events.NewCarMoved("car-1", 0, 0, 0, 1)
	
	store.Append(event1)
	store.Append(event2)
	
	if store.Count() != 2 {
		t.Fatalf("expected 2 events, got %d", store.Count())
	}
}

func TestStoreAll(t *testing.T) {
	store := New()
	
	store.Append(events.NewTickAdvanced(1))
	store.Append(events.NewCarMoved("car-1", 0, 0, 0, 1))
	
	all := store.All()
	
	if len(all) != 2 {
		t.Fatalf("expected 2 events, got %d", len(all))
	}
	
	if all[0].EventType() != "TickAdvanced" {
		t.Errorf("expected first event to be TickAdvanced, got %s", all[0].EventType())
	}
	
	if all[1].EventType() != "CarMoved" {
		t.Errorf("expected second event to be CarMoved, got %s", all[1].EventType())
	}
}

func TestStoreImmutability(t *testing.T) {
	store := New()
	
	store.Append(events.NewTickAdvanced(1))
	store.Append(events.NewTickAdvanced(2))
	
	// Get events
	events1 := store.All()
	
	// Mutate the returned slice
	events1[0] = events.NewTickAdvanced(999)
	
	// Get events again
	events2 := store.All()
	
	// Store should be unchanged
	tick1, ok := events2[0].(*events.TickAdvanced)
	if !ok {
		t.Fatal("expected TickAdvanced event")
	}
	
	if tick1.TickNumber == 999 {
		t.Error("store was mutated - immutability broken")
	}
	
	if tick1.TickNumber != 1 {
		t.Errorf("expected tick 1, got %d", tick1.TickNumber)
	}
}

func TestStoreSince(t *testing.T) {
	store := New()
	
	store.Append(events.NewTickAdvanced(1))
	store.Append(events.NewTickAdvanced(2))
	store.Append(events.NewTickAdvanced(3))
	
	since := store.Since(1)
	
	if len(since) != 2 {
		t.Fatalf("expected 2 events since index 1, got %d", len(since))
	}
	
	tick2, ok := since[0].(*events.TickAdvanced)
	if !ok || tick2.TickNumber != 2 {
		t.Error("expected first event to be tick 2")
	}
	
	tick3, ok := since[1].(*events.TickAdvanced)
	if !ok || tick3.TickNumber != 3 {
		t.Error("expected second event to be tick 3")
	}
}

func TestStoreClear(t *testing.T) {
	store := New()
	
	store.Append(events.NewTickAdvanced(1))
	store.Append(events.NewTickAdvanced(2))
	
	if store.Count() != 2 {
		t.Fatalf("expected 2 events before clear, got %d", store.Count())
	}
	
	store.Clear()
	
	if store.Count() != 0 {
		t.Errorf("expected 0 events after clear, got %d", store.Count())
	}
}