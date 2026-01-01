package engine

func Advance(state *State) {
	for _, car := range state.Cars {
		applyMovement(car)
	}
}
