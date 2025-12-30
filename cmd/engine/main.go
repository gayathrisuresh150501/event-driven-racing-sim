package main

import (
	"log"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
)

func main() {
	e := engine.New()
	log.Println("engine booted")
	e.Tick()
}
