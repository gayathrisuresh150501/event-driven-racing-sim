package main

import (
	"log"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

func main() {
	e := engine.New()
	log.Println("engine booted")
	
	// Subscribe to events to see them in real-time
	e.EventBus().Subscribe(func(event events.Event) {
		switch ev := event.(type) {
		case *events.TickAdvanced:
			log.Printf("Event: TickAdvanced(tick=%d)", ev.TickNumber)
		case *events.CarMoved:
			log.Printf("Event: CarMoved(car=%s, from=(%d,%d), to=(%d,%d))",
				ev.CarID, ev.OldX, ev.OldY, ev.NewX, ev.NewY)
		}
	})
	
	// Run a few ticks
	for i := 0; i < 3; i++ {
		e.Tick()
	}
	
	log.Printf("Total events stored: %d", e.EventStore().Count())
	log.Println("✅ Phase 2 complete: Event sourcing implemented!")
}