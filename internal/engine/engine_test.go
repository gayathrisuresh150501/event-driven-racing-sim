package engine

import "testing"

func TestDeterministicEngine(t *testing.T) {
	e1 := New()
	e2 := New()

	// Assert identical initial state
	c1, ok1 := e1.state.Cars["car-1"]
	c2, ok2 := e2.state.Cars["car-1"]

	if !ok1 || !ok2 {
		t.Fatalf("expected car-1 to exist in initial engine state")
	}

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

	f1, ok1 := e1.state.Cars["car-1"]
	f2, ok2 := e2.state.Cars["car-1"]

	if !ok1 || !ok2 {
		t.Fatalf("expected car-1 to exist after ticks")
	}

	if f1.Y != f2.Y {
		t.Fatalf(
			"engine is not deterministic: got %d and %d",
			f1.Y,
			f2.Y,
		)
	}
}
