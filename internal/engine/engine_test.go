package engine

import "testing"

func TestDeterministicEngine(t *testing.T) {
	e1 := New()
	e2 := New()

	// Assert identical initial state
	c1 := e1.state.Cars["car-1"]
	c2 := e2.state.Cars["car-1"]

	if c1.X != c2.X || c1.Y != c2.Y || c1.Speed != c2.Speed {
		t.Fatalf(
			"initial engine state differs: %+v vs %+v",
			c1,
			c2,
		)
	}

	for i := 0; i < 20; i++ {
		e1.Tick()
		e2.Tick()
	}

	if e1.state.Cars["car-1"].Y != e2.state.Cars["car-1"].Y {
		t.Fatalf(
			"engine is not deterministic: got %d and %d",
			e1.state.Cars["car-1"].Y,
			e2.state.Cars["car-1"].Y,
		)
	}
}
