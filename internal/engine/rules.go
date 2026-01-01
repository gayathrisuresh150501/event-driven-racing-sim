package engine

import "github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/model"

func applyMovement(car *model.Car) {
	car.Y += car.Speed
}
