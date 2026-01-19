package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/renderer"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║     Event-Driven Racing Simulator - Terminal Renderer      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create engine
	e := engine.New()

	// Create renderer (NEVER accesses engine state directly!)
	r := renderer.New(60, 15)

	// Subscribe renderer to engine events
	// This is the ONLY connection between engine and renderer
	unsubscribe := e.EventBus().Subscribe(func(event events.Event) {
		r.HandleEvent(event)
	})
	defer unsubscribe()

	// Create input handler
	input := renderer.NewInputHandler()

	// Initial render
	clearScreen()
	fmt.Println(r.Render())

	// Command loop
	fmt.Println("\nType 'help' for commands")

	for {
		cmd, err := input.ReadCommand()
		if err != nil {
			log.Printf("Error reading command: %v", err)
			continue
		}

		if cmd == nil {
			continue
		}

		// Handle commands
		if cmd.IsQuit() {
			fmt.Println("Goodbye!")
			break
		}

		if cmd.IsHelp() {
			fmt.Println(cmd.Help())
			continue
		}

		if cmd.IsTick() {
			// Execute single tick
			e.Tick()
		} else if cmd.IsTicks() {
			// Execute N ticks
			if len(cmd.Args) < 1 {
				fmt.Println("Usage: ticks N")
				continue
			}
			n, err := strconv.Atoi(cmd.Args[0])
			if err != nil || n < 1 {
				fmt.Println("Invalid number of ticks")
				continue
			}
			for i := 0; i < n; i++ {
				e.Tick()
			}
		} else {
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmd.Type)
			continue
		}

		// Re-render after command execution
		clearScreen()
		fmt.Println(r.Render())
		fmt.Printf("\nTotal events in store: %d\n", e.EventStore().Count())
	}
}

func clearScreen() {
	// ANSI escape code to clear screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")
}
