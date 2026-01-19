package eventbus

import (
	"testing"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

func TestBusPublishSubscribe(t *testing.T) {
	bus := New()
	
	var received []events.Event
	
	bus.Subscribe(func(e events.Event) {
		received = append(received, e)
	})
	
	event1 := events.NewTickAdvanced(1)
	event2 := events.NewCarMoved("car-1", 0, 0, 0, 1)
	
	bus.Publish(event1)
	bus.Publish(event2)
	
	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
	
	if received[0].EventType() != "TickAdvanced" {
		t.Errorf("expected first event to be TickAdvanced, got %s", received[0].EventType())
	}
	
	if received[1].EventType() != "CarMoved" {
		t.Errorf("expected second event to be CarMoved, got %s", received[1].EventType())
	}
}

func TestBusMultipleSubscribers(t *testing.T) {
	bus := New()
	
	count1 := 0
	count2 := 0
	
	bus.Subscribe(func(e events.Event) {
		count1++
	})
	
	bus.Subscribe(func(e events.Event) {
		count2++
	})
	
	bus.Publish(events.NewTickAdvanced(1))
	
	if count1 != 1 {
		t.Errorf("subscriber 1: expected 1 event, got %d", count1)
	}
	
	if count2 != 1 {
		t.Errorf("subscriber 2: expected 1 event, got %d", count2)
	}
}

func TestBusClear(t *testing.T) {
	bus := New()
	
	count := 0
	bus.Subscribe(func(e events.Event) {
		count++
	})
	
	bus.Publish(events.NewTickAdvanced(1))
	
	if count != 1 {
		t.Fatalf("expected count to be 1, got %d", count)
	}
	
	bus.Clear()
	bus.Publish(events.NewTickAdvanced(2))
	
	// Count should still be 1 because handlers were cleared
	if count != 1 {
		t.Errorf("expected count to remain 1 after clear, got %d", count)
	}
}