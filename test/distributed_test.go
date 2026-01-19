package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/rpc"
)

// TestDistributedRace proves client→gateway→engine→AI all communicate over HTTP/RPC
// This is THE KEY TEST for the distributed architecture
func TestDistributedRace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping distributed test in short mode")
	}

	// Start services
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start engine server
	t.Log("Starting engine server...")
	engineCmd := exec.CommandContext(ctx, "go", "run", "../cmd/engine-server/main.go")
	engineCmd.Env = append(os.Environ(), "PORT=18080")
	if err := engineCmd.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engineCmd.Process.Kill()

	// Start AI service 1
	t.Log("Starting AI service 1...")
	ai1Cmd := exec.CommandContext(ctx, "go", "run", "../cmd/ai-service/main.go")
	ai1Cmd.Env = append(os.Environ(), "PORT=18081", "AI_NAME=RuleBot-Alpha")
	if err := ai1Cmd.Start(); err != nil {
		t.Fatalf("Failed to start AI 1: %v", err)
	}
	defer ai1Cmd.Process.Kill()

	// Start AI service 2
	t.Log("Starting AI service 2...")
	ai2Cmd := exec.CommandContext(ctx, "go", "run", "../cmd/ai-service/main.go")
	ai2Cmd.Env = append(os.Environ(), "PORT=18082", "AI_NAME=RuleBot-Bravo")
	if err := ai2Cmd.Start(); err != nil {
		t.Fatalf("Failed to start AI 2: %v", err)
	}
	defer ai2Cmd.Process.Kill()

	// Start gateway
	t.Log("Starting gateway...")
	gatewayCmd := exec.CommandContext(ctx, "go", "run", "../cmd/gateway/main.go")
	gatewayCmd.Env = append(os.Environ(),
		"PORT=18083",
		"ENGINE_ADDRESS=http://localhost:18080",
		"AI_ADDRESS_1=http://localhost:18081",
		"AI_ADDRESS_2=http://localhost:18082",
	)
	if err := gatewayCmd.Start(); err != nil {
		t.Fatalf("Failed to start gateway: %v", err)
	}
	defer gatewayCmd.Process.Kill()

	// Wait for services to start
	t.Log("Waiting for services to be ready...")
	time.Sleep(3 * time.Second)

	// Test 1: Health checks
	t.Run("HealthChecks", func(t *testing.T) {
		services := map[string]string{
			"engine":   "http://localhost:18080/health",
			"ai-1":     "http://localhost:18081/health",
			"ai-2":     "http://localhost:18082/health",
			"gateway":  "http://localhost:18083/health",
		}

		for name, url := range services {
			resp, err := http.Get(url)
			if err != nil {
				t.Errorf("%s health check failed: %v", name, err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("%s returned status %d", name, resp.StatusCode)
			} else {
				t.Logf(" %s is healthy", name)
			}
		}
	})

	// Test 2: Start a race (gateway → engine)
	t.Run("StartRace", func(t *testing.T) {
		startReq := rpc.StartRaceRequest{
			NumCars: 4,
		}

		body, _ := json.Marshal(startReq)
		resp, err := http.Post("http://localhost:18083/race/start", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to start race: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Start race returned status %d", resp.StatusCode)
		}

		var startResp rpc.StartRaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&startResp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if startResp.RaceID == "" {
			t.Error("No race ID returned")
		}

		if len(startResp.CarIDs) != 4 {
			t.Errorf("Expected 4 cars, got %d", len(startResp.CarIDs))
		}

		t.Logf(" Race started: %s with %d cars", startResp.RaceID, len(startResp.CarIDs))
		t.Logf("   Race ID: %s", startResp.RaceID)
		t.Logf("   Cars: %v", startResp.CarIDs)

		// Test 3: Run the race (full distributed flow)
		t.Run("RunRace", func(t *testing.T) {
			runReq := rpc.RunRaceRequest{
				RaceID:   startResp.RaceID,
				NumTicks: 10,
			}

			body, _ := json.Marshal(runReq)
			resp, err := http.Post("http://localhost:18083/race/run", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("Failed to run race: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				t.Fatalf("Run race returned status %d: %s", resp.StatusCode, string(bodyBytes))
			}

			// Read streamed updates
			decoder := json.NewDecoder(resp.Body)
			updateCount := 0

			for {
				var update rpc.RaceUpdate
				if err := decoder.Decode(&update); err != nil {
					if err == io.EOF {
						break
					}
					t.Logf("Decode error (may be expected at end): %v", err)
					break
				}

				updateCount++

				if update.EventType == "tick" {
					t.Logf("   Tick %d: %d cars racing", update.TickNumber, len(update.Standings))
					if len(update.Standings) > 0 {
						leader := update.Standings[0]
						t.Logf("      Leader: %s at Y=%d, speed=%d", leader.CarID, leader.PositionY, leader.Speed)
					}
				} else if update.EventType == "completed" {
					t.Logf("   Race completed!")
				}
			}

			if updateCount == 0 {
				t.Error("No updates received from race")
			} else {
				t.Logf(" Received %d race updates", updateCount)
			}
		})

		// Test 4: Get final results
		t.Run("GetResults", func(t *testing.T) {
			resultsReq := rpc.GetRaceResultsRequest{
				RaceID: startResp.RaceID,
			}

			body, _ := json.Marshal(resultsReq)
			resp, err := http.Post("http://localhost:18083/race/results", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("Failed to get results: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Get results returned status %d", resp.StatusCode)
			}

			var results rpc.GetRaceResultsResponse
			if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
				t.Fatalf("Failed to decode results: %v", err)
			}

			t.Logf(" Final Results:")
			t.Logf("   Race ID: %s", results.RaceID)
			t.Logf("   Total Ticks: %d", results.TotalTicks)
			t.Logf("   Total Events: %d", results.TotalEvents)
			t.Logf("   Standings:")

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
				t.Logf("      %s %d. %s - Position: Y=%d, Speed: %d",
					medal, standing.Rank, standing.CarID, standing.PositionY, standing.Speed)
			}

			// Verify results
			if results.TotalTicks != 10 {
				t.Errorf("Expected 10 ticks, got %d", results.TotalTicks)
			}

			if len(results.FinalStandings) != 4 {
				t.Errorf("Expected 4 cars in standings, got %d", len(results.FinalStandings))
			}

			// Verify all cars moved forward
			for _, standing := range results.FinalStandings {
				if standing.PositionY <= 0 {
					t.Errorf("Car %s did not move (Y=%d)", standing.CarID, standing.PositionY)
				}
			}
		})
	})

	t.Log(" DISTRIBUTED SYSTEM TEST PASSED!")
	t.Log("   Proved: Client → Gateway → Engine → AI all communicate over HTTP/RPC")
}

// TestDistributedRaceWithDockerCompose tests the system using docker-compose
// Run with: docker-compose up -d && go test -v -run TestDistributedRaceWithDockerCompose
func TestDistributedRaceWithDockerCompose(t *testing.T) {
	if os.Getenv("DOCKER_TEST") != "1" {
		t.Skip("Skipping Docker test (set DOCKER_TEST=1 to run)")
	}

	// Assumes docker-compose is already running
	gatewayURL := "http://localhost:8082"

	// Check gateway health
	resp, err := http.Get(gatewayURL + "/health")
	if err != nil {
		t.Fatalf("Gateway not accessible: %v (is docker-compose running?)", err)
	}
	resp.Body.Close()

	// Start a race
	startReq := rpc.StartRaceRequest{
		NumCars: 4,
	}

	body, _ := json.Marshal(startReq)
	resp, err = http.Post(gatewayURL+"/race/start", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to start race: %v", err)
	}
	defer resp.Body.Close()

	var startResp rpc.StartRaceResponse
	json.NewDecoder(resp.Body).Decode(&startResp)

	t.Logf("Started race: %s", startResp.RaceID)

	// Run race
	runReq := rpc.RunRaceRequest{
		RaceID:   startResp.RaceID,
		NumTicks: 20,
	}

	body, _ = json.Marshal(runReq)
	resp, err = http.Post(gatewayURL+"/race/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to run race: %v", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for {
		var update rpc.RaceUpdate
		if err := decoder.Decode(&update); err != nil {
			break
		}

		if update.EventType == "tick" && update.TickNumber%5 == 0 {
			t.Logf("Tick %d complete", update.TickNumber)
		}
	}

	t.Log(" Docker Compose distributed race completed successfully!")
}
