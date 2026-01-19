package rpc

// Common RPC types for distributed services

// EngineService RPC types

type CreateRaceRequest struct {
	CarIDs []string `json:"car_ids"`
}

type CreateRaceResponse struct {
	RaceID  string   `json:"race_id"`
	CarIDs  []string `json:"car_ids"`
	Message string   `json:"message"`
}

type TickRequest struct {
	RaceID   string `json:"race_id"`
	NumTicks int    `json:"num_ticks"`
}

type TickResponse struct {
	CurrentTick int        `json:"current_tick"`
	Cars        []CarState `json:"cars"`
}

type GetStateRequest struct {
	RaceID string `json:"race_id"`
}

type GetStateResponse struct {
	CurrentTick int        `json:"current_tick"`
	Cars        []CarState `json:"cars"`
}

type RegisterAIRequest struct {
	RaceID           string `json:"race_id"`
	CarID            string `json:"car_id"`
	AIServiceAddress string `json:"ai_service_address"`
}

type RegisterAIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CarState struct {
	ID    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Speed int    `json:"speed"`
}

// AIService RPC types

type DecideRequest struct {
	Car        CarState   `json:"car"`
	AllCars    []CarState `json:"all_cars"`
	TickNumber int        `json:"tick_number"`
	TimeoutMs  int64      `json:"timeout_ms"`
}

type DecideResponse struct {
	CarID           string  `json:"car_id"`
	SuggestedDeltaY int     `json:"suggested_delta_y"`
	Confidence      float64 `json:"confidence"`
	Error           string  `json:"error,omitempty"`
}

type GetAIInfoRequest struct{}

type GetAIInfoResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// GatewayService RPC types

type StartRaceRequest struct {
	NumCars   int        `json:"num_cars"`
	AIConfigs []AIConfig `json:"ai_configs"`
}

type AIConfig struct {
	CarID            string `json:"car_id"`
	AIType           string `json:"ai_type"`
	AIServiceAddress string `json:"ai_service_address,omitempty"`
}

type StartRaceResponse struct {
	RaceID  string   `json:"race_id"`
	CarIDs  []string `json:"car_ids"`
	Message string   `json:"message"`
}

type RunRaceRequest struct {
	RaceID   string `json:"race_id"`
	NumTicks int    `json:"num_ticks"`
}

type RaceUpdate struct {
	TickNumber int           `json:"tick_number"`
	Standings  []CarStanding `json:"standings"`
	EventType  string        `json:"event_type"` // "tick", "completed"
}

type CarStanding struct {
	Rank      int    `json:"rank"`
	CarID     string `json:"car_id"`
	PositionY int    `json:"position_y"`
	Speed     int    `json:"speed"`
	AIName    string `json:"ai_name"`
}

type GetRaceResultsRequest struct {
	RaceID string `json:"race_id"`
}

type GetRaceResultsResponse struct {
	RaceID         string        `json:"race_id"`
	TotalTicks     int           `json:"total_ticks"`
	FinalStandings []CarStanding `json:"final_standings"`
	TotalEvents    int           `json:"total_events"`
}
