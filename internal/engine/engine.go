package engine

type Engine struct {
	state *State
}

func New() *Engine {
	return &Engine{
		state: NewState(),
	}
}

func (e *Engine) Tick() {
	Advance(e.state)
}
