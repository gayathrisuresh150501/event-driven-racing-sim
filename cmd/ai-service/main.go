package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/rpc"
)

type AIService struct {
	name        string
	version     string
	description string
}

func NewAIService(name, description string) *AIService {
	return &AIService{
		name:        name,
		version:     "1.0.0",
		description: description,
	}
}

func (s *AIService) handleDecide(w http.ResponseWriter, r *http.Request) {
	var req rpc.DecideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Implement RuleBot logic:
	// 1. Base speed increases every 10 ticks (1→2→3→4→5, capped at 5)
	// 2. If more than 10 units behind leader, boost speed by 1

	desiredSpeed := 1 + (req.TickNumber / 10)
	if desiredSpeed > 5 {
		desiredSpeed = 5
	}

	// Check if any car is ahead
	maxY := req.Car.Y
	for _, otherCar := range req.AllCars {
		if otherCar.ID != req.Car.ID && otherCar.Y > maxY {
			maxY = otherCar.Y
		}
	}

	// If we're more than 10 units behind the leader, boost speed
	if maxY-req.Car.Y > 10 {
		desiredSpeed++
		if desiredSpeed > 5 {
			desiredSpeed = 5
		}
	}

	resp := rpc.DecideResponse{
		CarID:           req.Car.ID,
		SuggestedDeltaY: desiredSpeed,
		Confidence:      0.8,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Printf("[AI] Decide for %s at tick %d: speed=%d", req.Car.ID, req.TickNumber, desiredSpeed)
}

func (s *AIService) handleGetInfo(w http.ResponseWriter, r *http.Request) {
	resp := rpc.GetAIInfoResponse{
		Name:        s.name,
		Version:     s.version,
		Description: s.description,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *AIService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	aiName := os.Getenv("AI_NAME")
	if aiName == "" {
		aiName = "RuleBot"
	}

	service := NewAIService(aiName, "Simple rule-based racing AI")

	http.HandleFunc("/health", service.handleHealth)
	http.HandleFunc("/decide", service.handleDecide)
	http.HandleFunc("/info", service.handleGetInfo)

	addr := ":" + port
	log.Printf("[AI] Starting AI service '%s' on %s", aiName, addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
