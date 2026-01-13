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
// If timestamps is provided, the first one is used. Otherwise, time.Now() is used.
func NewTickAdvanced(tickNumber int, timestamps ...time.Time) *TickAdvanced {
	ts := time.Now()
	if len(timestamps) > 0 {
		ts = timestamps[0]
	}
	return &TickAdvanced{
		BaseEvent: BaseEvent{
			Type: "TickAdvanced",
			Time: ts,
		},
		TickNumber: tickNumber,
	}
}

// NewCarMoved creates a new CarMoved event
// If timestamps is provided, the first one is used. Otherwise, time.Now() is used.
func NewCarMoved(carID string, oldX, oldY, newX, newY int, timestamps ...time.Time) *CarMoved {
	ts := time.Now()
	if len(timestamps) > 0 {
		ts = timestamps[0]
	}
	return &CarMoved{
		BaseEvent: BaseEvent{
			Type: "CarMoved",
			Time: ts,
		},
		CarID: carID,
		OldX:  oldX,
		OldY:  oldY,
		NewX:  newX,
		NewY:  newY,
	}
}
