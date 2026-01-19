package engineserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/internal/engine"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"
)

// Server implements the engine gRPC service
type Server struct {
	mu         sync.RWMutex
	engines    map[string]*engine.Engine
	aiClients  map[string]map[string]AIClient // raceID -> carID -> AI client
	eventChans map[string][]chan events.Event  // raceID -> subscribers
}

// AIClient represents a remote AI service
type AIClient interface {
	Decide(ctx context.Context, car *model.Car, allCars []*model.Car, tickNumber int) (Decision, error)
	Name() string
}

// Decision represents an AI's suggested action
type Decision struct {
	CarID           string
	SuggestedDeltaY int
	Confidence      float64
}

// NewServer creates a new engine server
func NewServer() *Server {
	return &Server{
		engines:    make(map[string]*engine.Engine),
		aiClients:  make(map[string]map[string]AIClient),
		eventChans: make(map[string][]chan events.Event),
	}
}

// CreateRace initializes a new race
func (s *Server) CreateRace(ctx context.Context, carIDs []string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	raceID := fmt.Sprintf("race-%d", len(s.engines)+1)
	eng := engine.New()

	// Add additional cars beyond the default car-1
	for i, carID := range carIDs {
		if i == 0 {
			// car-1 already exists in default state
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
	s.aiClients[raceID] = make(map[string]AIClient)
	s.eventChans[raceID] = make([]chan events.Event, 0)

	return raceID, nil
}

// Tick advances the simulation
func (s *Server) Tick(ctx context.Context, raceID string, numTicks int) (int, []*model.Car, error) {
	s.mu.RLock()
	eng, exists := s.engines[raceID]
	s.mu.RUnlock()

	if !exists {
		return 0, nil, fmt.Errorf("race not found: %s", raceID)
	}

	for i := 0; i < numTicks; i++ {
		eng.Tick()
	}

	// Get current state
	state := eng.GetState()
	cars := make([]*model.Car, 0, len(state.Cars))
	for _, car := range state.Cars {
		cars = append(cars, car)
	}

	return state.TickNumber, cars, nil
}

// GetState returns the current race state
func (s *Server) GetState(ctx context.Context, raceID string) (int, []*model.Car, error) {
	s.mu.RLock()
	eng, exists := s.engines[raceID]
	s.mu.RUnlock()

	if !exists {
		return 0, nil, fmt.Errorf("race not found: %s", raceID)
	}

	state := eng.GetState()
	cars := make([]*model.Car, 0, len(state.Cars))
	for _, car := range state.Cars {
		cars = append(cars, car)
	}

	return state.TickNumber, cars, nil
}

// RegisterAI registers an AI client for a specific car
func (s *Server) RegisterAI(ctx context.Context, raceID, carID string, aiClient AIClient) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.engines[raceID]; !exists {
		return fmt.Errorf("race not found: %s", raceID)
	}

	if s.aiClients[raceID] == nil {
		s.aiClients[raceID] = make(map[string]AIClient)
	}

	s.aiClients[raceID][carID] = aiClient

	// Register AI with engine
	eng := s.engines[raceID]
	aiAdapter := &remoteAIAdapter{
		client: aiClient,
	}
	eng.RegisterAI(carID, aiAdapter)

	return nil
}

// GetEventStore returns the event store for a race
func (s *Server) GetEventStore(raceID string) ([]events.Event, error) {
	s.mu.RLock()
	eng, exists := s.engines[raceID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("race not found: %s", raceID)
	}

	return eng.EventStore().All(), nil
}
