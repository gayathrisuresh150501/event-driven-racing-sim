package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/rpc"
)

type Gateway struct {
	engineAddress string
	aiAddresses   []string
	client        *http.Client
}

func NewGateway(engineAddr string, aiAddrs []string) *Gateway {
	return &Gateway{
		engineAddress: engineAddr,
		aiAddresses:   aiAddrs,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *Gateway) handleStartRace(w http.ResponseWriter, r *http.Request) {
	var req rpc.StartRaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate car IDs
	carIDs := make([]string, req.NumCars)
	for i := 0; i < req.NumCars; i++ {
		carIDs[i] = fmt.Sprintf("car-%d", i+1)
	}

	// Create race on engine
	createReq := rpc.CreateRaceRequest{
		CarIDs: carIDs,
	}

	createResp, err := g.callEngine("/race/create", createReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create race: %v", err), http.StatusInternalServerError)
		return
	}

	var raceResp rpc.CreateRaceResponse
	if err := json.Unmarshal(createResp, &raceResp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Register AI services for each car
	for i, carID := range carIDs {
		aiAddr := g.aiAddresses[i%len(g.aiAddresses)]

		regReq := rpc.RegisterAIRequest{
			RaceID:           raceResp.RaceID,
			CarID:            carID,
			AIServiceAddress: aiAddr,
		}

		_, err := g.callEngine("/race/register-ai", regReq)
		if err != nil {
			log.Printf("[GATEWAY] Warning: failed to register AI for %s: %v", carID, err)
		}
	}

	resp := rpc.StartRaceResponse{
		RaceID:  raceResp.RaceID,
		CarIDs:  carIDs,
		Message: fmt.Sprintf("Race %s started with %d cars", raceResp.RaceID, len(carIDs)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("[GATEWAY] Started race %s with %d cars", raceResp.RaceID, len(carIDs))
}

func (g *Gateway) handleRunRace(w http.ResponseWriter, r *http.Request) {
	var req rpc.RunRaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	encoder := json.NewEncoder(w)

	// Run race tick by tick and stream updates
	for tick := 1; tick <= req.NumTicks; tick++ {
		tickReq := rpc.TickRequest{
			RaceID:   req.RaceID,
			NumTicks: 1,
		}

		tickResp, err := g.callEngine("/race/tick", tickReq)
		if err != nil {
			log.Printf("[GATEWAY] Error on tick %d: %v", tick, err)
			break
		}

		var state rpc.TickResponse
		if err := json.Unmarshal(tickResp, &state); err != nil {
			log.Printf("[GATEWAY] Error unmarshaling tick response: %v", err)
			break
		}

		// Calculate standings
		standings := g.calculateStandings(state.Cars)

		update := rpc.RaceUpdate{
			TickNumber: state.CurrentTick,
			Standings:  standings,
			EventType:  "tick",
		}

		if err := encoder.Encode(update); err != nil {
			log.Printf("[GATEWAY] Error encoding update: %v", err)
			break
		}

		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		// Small delay between ticks for visibility
		time.Sleep(50 * time.Millisecond)
	}

	// Send final update
	finalUpdate := rpc.RaceUpdate{
		EventType: "completed",
	}
	encoder.Encode(finalUpdate)

	log.Printf("[GATEWAY] Completed race %s after %d ticks", req.RaceID, req.NumTicks)
}

func (g *Gateway) handleGetRaceResults(w http.ResponseWriter, r *http.Request) {
	var req rpc.GetRaceResultsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stateReq := rpc.GetStateRequest{
		RaceID: req.RaceID,
	}

	stateResp, err := g.callEngine("/race/state", stateReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get state: %v", err), http.StatusInternalServerError)
		return
	}

	var state rpc.GetStateResponse
	if err := json.Unmarshal(stateResp, &state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	standings := g.calculateStandings(state.Cars)

	resp := rpc.GetRaceResultsResponse{
		RaceID:         req.RaceID,
		TotalTicks:     state.CurrentTick,
		FinalStandings: standings,
		TotalEvents:    state.CurrentTick * (1 + len(state.Cars)), // Estimate
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (g *Gateway) calculateStandings(cars []rpc.CarState) []rpc.CarStanding {
	// Sort cars by position Y (descending)
	sort.Slice(cars, func(i, j int) bool {
		return cars[i].Y > cars[j].Y
	})

	standings := make([]rpc.CarStanding, len(cars))
	for i, car := range cars {
		standings[i] = rpc.CarStanding{
			Rank:      i + 1,
			CarID:     car.ID,
			PositionY: car.Y,
			Speed:     car.Speed,
			AIName:    "RuleBot", // Could fetch from AI service
		}
	}

	return standings
}

func (g *Gateway) callEngine(path string, reqBody interface{}) ([]byte, error) {
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", g.engineAddress+path, io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("engine returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	engineAddr := os.Getenv("ENGINE_ADDRESS")
	if engineAddr == "" {
		engineAddr = "http://localhost:8080"
	}

	aiAddr1 := os.Getenv("AI_ADDRESS_1")
	if aiAddr1 == "" {
		aiAddr1 = "http://localhost:8081"
	}

	aiAddr2 := os.Getenv("AI_ADDRESS_2")
	if aiAddr2 == "" {
		aiAddr2 = "http://localhost:8081"
	}

	gateway := NewGateway(engineAddr, []string{aiAddr1, aiAddr2})

	http.HandleFunc("/health", gateway.handleHealth)
	http.HandleFunc("/race/start", gateway.handleStartRace)
	http.HandleFunc("/race/run", gateway.handleRunRace)
	http.HandleFunc("/race/results", gateway.handleGetRaceResults)

	addr := ":" + port
	log.Printf("[GATEWAY] Starting gateway on %s", addr)
	log.Printf("[GATEWAY] Engine: %s", engineAddr)
	log.Printf("[GATEWAY] AI Services: %v", []string{aiAddr1, aiAddr2})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
