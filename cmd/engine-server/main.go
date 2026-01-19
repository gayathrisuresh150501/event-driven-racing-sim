package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/rpc"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

type EngineServer struct {
	mu        sync.RWMutex
	engines   map[string]*engine.Engine
	aiClients map[string]map[string]*HTTPAIClient // raceID -> carID -> AI client
}

func NewEngineServer() *EngineServer {
	return &EngineServer{
		engines:   make(map[string]*engine.Engine),
		aiClients: make(map[string]map[string]*HTTPAIClient),
	}
}

func (s *EngineServer) handleCreateRace(w http.ResponseWriter, r *http.Request) {
	var req rpc.CreateRaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	raceID := fmt.Sprintf("race-%d", len(s.engines)+1)
	eng := engine.New()

	// Add additional cars
	for i, carID := range req.CarIDs {
		if i == 0 && carID == "car-1" {
			// car-1 already exists
			continue
		}
		eng.GetState().Cars[carID] = &model.Car{
			ID:    carID,
			X:     0,
			Y:     0,
			Speed: 1,
		}
	}

	s.engines[raceID] = eng
	s.aiClients[raceID] = make(map[string]*HTTPAIClient)

	resp := rpc.CreateRaceResponse{
		RaceID:  raceID,
		CarIDs:  req.CarIDs,
		Message: fmt.Sprintf("Race created with %d cars", len(req.CarIDs)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("[ENGINE] Created race %s with %d cars", raceID, len(req.CarIDs))
}

func (s *EngineServer) handleTick(w http.ResponseWriter, r *http.Request) {
	var req rpc.TickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	eng, exists := s.engines[req.RaceID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "race not found", http.StatusNotFound)
		return
	}

	if req.NumTicks <= 0 {
		req.NumTicks = 1
	}

	for i := 0; i < req.NumTicks; i++ {
		eng.Tick()
	}

	state := eng.GetState()
	cars := make([]rpc.CarState, 0, len(state.Cars))
	for _, car := range state.Cars {
		cars = append(cars, rpc.CarState{
			ID:    car.ID,
			X:     car.X,
			Y:     car.Y,
			Speed: car.Speed,
		})
	}

	resp := rpc.TickResponse{
		CurrentTick: state.TickNumber,
		Cars:        cars,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("[ENGINE] Tick %s: tick=%d, cars=%d", req.RaceID, state.TickNumber, len(cars))
}

func (s *EngineServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	var req rpc.GetStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	eng, exists := s.engines[req.RaceID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "race not found", http.StatusNotFound)
		return
	}

	state := eng.GetState()
	cars := make([]rpc.CarState, 0, len(state.Cars))
	for _, car := range state.Cars {
		cars = append(cars, rpc.CarState{
			ID:    car.ID,
			X:     car.X,
			Y:     car.Y,
			Speed: car.Speed,
		})
	}

	resp := rpc.GetStateResponse{
		CurrentTick: state.TickNumber,
		Cars:        cars,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *EngineServer) handleRegisterAI(w http.ResponseWriter, r *http.Request) {
	var req rpc.RegisterAIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.engines[req.RaceID]
	if !exists {
		http.Error(w, "race not found", http.StatusNotFound)
		return
	}

	// Create HTTP AI client
	aiClient := NewHTTPAIClient(req.AIServiceAddress)

	// Get AI info
	info, err := aiClient.GetInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to contact AI service: %v", err), http.StatusBadRequest)
		return
	}

	// Store client
	if s.aiClients[req.RaceID] == nil {
		s.aiClients[req.RaceID] = make(map[string]*HTTPAIClient)
	}
	s.aiClients[req.RaceID][req.CarID] = aiClient

	resp := rpc.RegisterAIResponse{
		Success: true,
		Message: fmt.Sprintf("Registered AI '%s' for car %s", info.Name, req.CarID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("[ENGINE] Registered AI '%s' for %s in race %s", info.Name, req.CarID, req.RaceID)
}

func (s *EngineServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewEngineServer()

	http.HandleFunc("/health", server.handleHealth)
	http.HandleFunc("/race/create", server.handleCreateRace)
	http.HandleFunc("/race/tick", server.handleTick)
	http.HandleFunc("/race/state", server.handleGetState)
	http.HandleFunc("/race/register-ai", server.handleRegisterAI)

	addr := ":" + port
	log.Printf("[ENGINE] Starting engine server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// HTTPAIClient is a client for remote HTTP AI services
type HTTPAIClient struct {
	address string
	client  *http.Client
	name    string
}

func NewHTTPAIClient(address string) *HTTPAIClient {
	return &HTTPAIClient{
		address: address,
		client: &http.Client{
			Timeout: 150 * time.Millisecond,
		},
	}
}

func (c *HTTPAIClient) Decide(ctx context.Context, car rpc.CarState, allCars []rpc.CarState, tickNumber int) (*rpc.DecideResponse, error) {
	req := rpc.DecideRequest{
		Car:        car,
		AllCars:    allCars,
		TickNumber: tickNumber,
		TimeoutMs:  100,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.address+"/decide", io.NopCloser(bytes.NewReader(body)))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var decideResp rpc.DecideResponse
	if err := json.NewDecoder(resp.Body).Decode(&decideResp); err != nil {
		return nil, err
	}

	if decideResp.Error != "" {
		return nil, fmt.Errorf("%s", decideResp.Error)
	}

	return &decideResp, nil
}

func (c *HTTPAIClient) GetInfo() (*rpc.GetAIInfoResponse, error) {
	resp, err := c.client.Post(c.address+"/info", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info rpc.GetAIInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	c.name = info.Name
	return &info, nil
}
