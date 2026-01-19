package main

import (
	"fmt"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/ai"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/renderer"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

func main() {
	fmt.Println("🏁 AI Racing Simulation Demo")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("Key Principle: AI suggests, engine decides. Engine NEVER waits for AI.")
	fmt.Println()

	// Create the engine
	e := engine.New()

	// Add multiple cars to the race
	e.GetState().Cars["car-2"] = &model.Car{ID: "car-2", X: 0, Y: 0, Speed: 1}
	e.GetState().Cars["car-3"] = &model.Car{ID: "car-3", X: 0, Y: 0, Speed: 1}
	e.GetState().Cars["car-4"] = &model.Car{ID: "car-4", X: 0, Y: 0, Speed: 1}

	// Register AI agents for each car (all with 100ms timeout protection)
	e.RegisterAI("car-1", ai.NewTimeoutWrapper(ai.NewRuleBot("Alpha"), 100*time.Millisecond, 1))
	e.RegisterAI("car-2", ai.NewTimeoutWrapper(ai.NewRuleBot("Bravo"), 100*time.Millisecond, 1))
	e.RegisterAI("car-3", ai.NewTimeoutWrapper(ai.NewRuleBot("Charlie"), 100*time.Millisecond, 1))
	e.RegisterAI("car-4", ai.NewTimeoutWrapper(ai.NewRuleBot("Delta"), 100*time.Millisecond, 1))

	// Create a renderer to display the race
	r := renderer.New(80, 24)
	e.EventBus().Subscribe(func(event events.Event) {
		r.HandleEvent(event)
	})

	fmt.Println("Starting race with 4 AI-controlled cars...")
	fmt.Println()

	// List the AI agents
	for carID, agent := range e.GetAIAgents() {
		fmt.Printf("  %s controlled by AI: %s\n", carID, agent.Name())
	}
	fmt.Println()

	// Run the race for 30 ticks
	const numTicks = 30

	fmt.Printf("Running %d ticks...\n", numTicks)
	fmt.Println()

	start := time.Now()

	// Run in batches and display progress
	for i := 0; i < numTicks; i++ {
		e.Tick()

		// Display every 5 ticks
		if (i+1)%5 == 0 {
			fmt.Printf("=== After Tick %d ===\n", i+1)
			displayStandings(e)
			fmt.Println()
		}
	}

	elapsed := time.Since(start)

	// Final results
	fmt.Println("🏆 FINAL RESULTS 🏆")
	fmt.Println("===================")
	fmt.Println()

	displayStandings(e)

	fmt.Println()
	fmt.Printf("Race completed in %v\n", elapsed)
	fmt.Printf("Total events generated: %d\n", e.EventStore().Count())
	fmt.Printf("Average time per tick: %v\n", elapsed/numTicks)
	fmt.Println()
	fmt.Println("✅ All AI agents made autonomous decisions!")
	fmt.Println("✅ Engine never blocked waiting for AI!")
}

func displayStandings(e *engine.Engine) {
	// Get all cars and sort by position
	type carInfo struct {
		id       string
		position int
		speed    int
		aiName   string
	}

	var standings []carInfo
	for id, car := range e.GetState().Cars {
		ai := e.GetAIAgents()[id]
		aiName := "None"
		if ai != nil {
			aiName = ai.Name()
		}

		standings = append(standings, carInfo{
			id:       id,
			position: car.Y,
			speed:    car.Speed,
			aiName:   aiName,
		})
	}

	// Sort by position (descending)
	for i := 0; i < len(standings); i++ {
		for j := i + 1; j < len(standings); j++ {
			if standings[j].position > standings[i].position {
				standings[i], standings[j] = standings[j], standings[i]
			}
		}
	}

	// Display standings
	for rank, info := range standings {
		medal := ""
		switch rank {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = "  "
		}

		fmt.Printf("%s %d. %s - Position: Y=%d, Speed: %d (AI: %s)\n",
			medal, rank+1, info.id, info.position, info.speed, info.aiName)
	}
}
