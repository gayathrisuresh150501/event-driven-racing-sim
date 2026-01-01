package engine

import "testing"

func TestDeterministicEngine(t *testing.T) {
	e1 := New()
	e2 := New()

	for i := 0; i < 20; i++ {
		e1.Tick()
		e2.Tick()
	}

	car1 := e1.state.Cars["car-1"]
	car2 := e2.state.Cars["car-1"]

	if car1.Y != car2.Y {
		t.Fatalf(
			"engine is not deterministic: got %d and %d",
			car1.Y,
			car2.Y,
		)
	}
}
