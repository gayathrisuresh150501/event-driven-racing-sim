package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/rpc"
)

func main() {
	gatewayURL := flag.String("gateway", "http://localhost:8082", "Gateway URL")
	numCars := flag.Int("cars", 4, "Number of cars")
	numTicks := flag.Int("ticks", 20, "Number of ticks to run")
	flag.Parse()

	fmt.Println("🏁 Distributed Racing Client")
	fmt.Println("============================")
	fmt.Printf("Gateway: %s\n", *gatewayURL)
	fmt.Printf("Cars: %d\n", *numCars)
	fmt.Printf("Ticks: %d\n", *numTicks)
	fmt.Println()

	client := &http.Client{}

	// Step 1: Check gateway health
	fmt.Println("Checking gateway health...")
	resp, err := client.Get(*gatewayURL + "/health")
	if err != nil {
		log.Fatalf("Gateway not accessible: %v", err)
	}
	resp.Body.Close()
	fmt.Println(" Gateway is healthy")
	fmt.Println()

	// Step 2: Start a race
	fmt.Printf("Starting race with %d cars...\n", *numCars)
	startReq := rpc.StartRaceRequest{
		NumCars: *numCars,
	}

	body, _ := json.Marshal(startReq)
	resp, err = client.Post(*gatewayURL+"/race/start", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Failed to start race: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Start race failed: %s", string(bodyBytes))
	}

	var startResp rpc.StartRaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&startResp); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	fmt.Printf(" Race started: %s\n", startResp.RaceID)
	fmt.Printf("   Cars: %v\n", startResp.CarIDs)
	fmt.Println()

	// Step 3: Run the race
	fmt.Printf("Running race for %d ticks...\n", *numTicks)
	runReq := rpc.RunRaceRequest{
		RaceID:   startResp.RaceID,
		NumTicks: *numTicks,
	}

	body, _ = json.Marshal(runReq)
	resp, err = client.Post(*gatewayURL+"/race/run", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Failed to run race: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Run race failed: %s", string(bodyBytes))
	}

	// Stream race updates
	decoder := json.NewDecoder(resp.Body)

	for {
		var update rpc.RaceUpdate
		if err := decoder.Decode(&update); err != nil {
			if err == io.EOF {
				break
			}
			// Ignore decode errors at end of stream
			break
		}

		if update.EventType == "tick" {
			// Display every 5 ticks
			if update.TickNumber%5 == 0 || update.TickNumber == 1 {
				fmt.Printf("\n=== Tick %d ===\n", update.TickNumber)
				for _, standing := range update.Standings {
					medal := ""
					switch standing.Rank {
					case 1:
						medal = "🥇"
					case 2:
						medal = "🥈"
					case 3:
						medal = "🥉"
					default:
						medal = "  "
					}
					fmt.Printf("%s %d. %s - Y=%d, Speed=%d\n",
						medal, standing.Rank, standing.CarID, standing.PositionY, standing.Speed)
				}
			}
		} else if update.EventType == "completed" {
			fmt.Println("\n🏁 Race completed!")
		}
	}

	fmt.Println()

	// Step 4: Get final results
	fmt.Println("Getting final results...")
	resultsReq := rpc.GetRaceResultsRequest{
		RaceID: startResp.RaceID,
	}

	body, _ = json.Marshal(resultsReq)
	resp, err = client.Post(*gatewayURL+"/race/results", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("Failed to get results: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Get results failed: %s", string(bodyBytes))
	}

	var results rpc.GetRaceResultsResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		log.Fatalf("Failed to decode results: %v", err)
	}

	fmt.Println()
	fmt.Println("🏆 FINAL RESULTS 🏆")
	fmt.Println("===================")
	fmt.Println()

	for _, standing := range results.FinalStandings {
		medal := ""
		switch standing.Rank {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		default:
			medal = "  "
		}
		fmt.Printf("%s %d. %s - Position: Y=%d, Speed: %d (AI: %s)\n",
			medal, standing.Rank, standing.CarID, standing.PositionY, standing.Speed, standing.AIName)
	}

	fmt.Println()
	fmt.Printf("Race ID: %s\n", results.RaceID)
	fmt.Printf("Total Ticks: %d\n", results.TotalTicks)
	fmt.Printf("Total Events: %d\n", results.TotalEvents)
	fmt.Println()
	fmt.Println(" Distributed race completed successfully!")
	fmt.Println("   Proved: Client → Gateway → Engine → AI all communicate over network")

	os.Exit(0)
}
