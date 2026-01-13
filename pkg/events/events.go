package events

import "time"

// Event is the interface that all domain events must implement
type Event interface {
	EventType() string
	Timestamp() time.Time
}

// BaseEvent contains common fields for all events
type BaseEvent struct {
	Type string    `json:"type"`
	Time time.Time `json:"timestamp"`
}

func (e BaseEvent) EventType() string {
	return e.Type
}

func (e BaseEvent) Timestamp() time.Time {
	return e.Time
}

// TickAdvanced represents the engine advancing by one tick
type TickAdvanced struct {
	BaseEvent
	TickNumber int `json:"tickNumber"`
}

// CarMoved represents a car changing position
type CarMoved struct {
	BaseEvent
	CarID string `json:"carId"`
	OldX  int    `json:"oldX"`
	OldY  int    `json:"oldY"`
	NewX  int    `json:"newX"`
	NewY  int    `json:"newY"`
}

// NewTickAdvanced creates a new TickAdvanced event
func NewTickAdvanced(tickNumber int) *TickAdvanced {
	return &TickAdvanced{
		BaseEvent: BaseEvent{
			Type: "TickAdvanced",
			Time: time.Now(),
		},
		TickNumber: tickNumber,
	}
}

// NewCarMoved creates a new CarMoved event
func NewCarMoved(carID string, oldX, oldY, newX, newY int) *CarMoved {
	return &CarMoved{
		BaseEvent: BaseEvent{
			Type: "CarMoved",
			Time: time.Now(),
		},
		CarID: carID,
		OldX:  oldX,
		OldY:  oldY,
		NewX:  newX,
		NewY:  newY,
	}
}