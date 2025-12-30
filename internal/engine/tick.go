package engine

func Advance(state *State) {
	for _, car := range state.Cars {
		car.Y += car.Speed
	}
}
