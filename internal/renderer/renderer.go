package renderer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gayathrisuresh150501/event-driven-racing-sim/pkg/events"
)

// Renderer displays the racing sim in a terminal
// CRITICAL: It NEVER reads engine state directly - only subscribes to events
type Renderer struct {
	state  *State
	width  int
	height int
}

// New creates a new renderer
func New(width, height int) *Renderer {
	return &Renderer{
		state:  NewState(),
		width:  width,
		height: height,
	}
}

// HandleEvent processes an event and updates the renderer's state
// This is the ONLY way the renderer learns about the world
func (r *Renderer) HandleEvent(event events.Event) {
	r.state.ApplyEvent(event)
}

// Render produces a string representation of the current state
// This demonstrates the renderer works entirely from event-sourced state
func (r *Renderer) Render() string {
	var buf strings.Builder

	// Title bar
	buf.WriteString("╔")
	buf.WriteString(strings.Repeat("═", r.width-2))
	buf.WriteString("╗\n")

	// HUD
	buf.WriteString(r.renderHUD())

	// Track divider
	buf.WriteString("╠")
	buf.WriteString(strings.Repeat("═", r.width-2))
	buf.WriteString("╣\n")

	// Track view (pseudo-3D)
	buf.WriteString(r.renderTrack())

	// Bottom border
	buf.WriteString("╚")
	buf.WriteString(strings.Repeat("═", r.width-2))
	buf.WriteString("╝\n")

	return buf.String()
}

// renderHUD creates the heads-up display with tick, car count, and car stats
func (r *Renderer) renderHUD() string {
	var buf strings.Builder

	// First line: Tick and Car count
	line1 := fmt.Sprintf(" Tick: %d | Cars: %d", r.state.TickNumber, len(r.state.Cars))
	buf.WriteString("║")
	buf.WriteString(line1)
	buf.WriteString(strings.Repeat(" ", r.width-len(line1)-2))
	buf.WriteString("║\n")

	// Car details (sorted by ID for determinism)
	carIDs := make([]string, 0, len(r.state.Cars))
	for id := range r.state.Cars {
		carIDs = append(carIDs, id)
	}
	sort.Strings(carIDs)

	for _, id := range carIDs {
		car := r.state.Cars[id]
		line := fmt.Sprintf(" %s: Pos(%d,%d) Speed:%d", car.ID, car.X, car.Y, car.Speed)
		buf.WriteString("║")
		buf.WriteString(line)
		buf.WriteString(strings.Repeat(" ", r.width-len(line)-2))
		buf.WriteString("║\n")
	}

	return buf.String()
}

// renderTrack creates a pseudo-3D perspective view of the track
func (r *Renderer) renderTrack() string {
	var buf strings.Builder

	trackHeight := 10
	trackWidth := r.width - 2

	for row := 0; row < trackHeight; row++ {
		buf.WriteString("║")

		// Calculate perspective depth (row 0 is far, row 9 is near)
		depth := float64(row) / float64(trackHeight-1)

		// Track edges narrow at the distance (perspective effect)
		leftMargin := int((1.0 - depth) * float64(trackWidth) / 4)
		rightMargin := leftMargin
		roadWidth := trackWidth - leftMargin - rightMargin

		// Left margin (grass)
		buf.WriteString(strings.Repeat("░", leftMargin))

		// Road markings
		if roadWidth > 0 {
			// Road edges
			buf.WriteString("│")

			// Center line dashes (every other row for depth effect)
			if roadWidth > 2 {
				roadInner := roadWidth - 2
				centerPos := roadInner / 2

				for i := 0; i < roadInner; i++ {
					if i == centerPos && row%2 == 0 {
						buf.WriteString("│") // Center line
					} else {
						// Check if any car is at this position
						carChar := r.findCarAt(row, leftMargin+1+i, trackWidth, trackHeight)
						if carChar != "" {
							buf.WriteString(carChar)
						} else {
							buf.WriteString(" ")
						}
					}
				}
			}

			buf.WriteString("│")
		}

		// Right margin (grass)
		buf.WriteString(strings.Repeat("░", rightMargin))

		// Pad remaining width
		currentLen := leftMargin + roadWidth + rightMargin
		if currentLen < trackWidth {
			buf.WriteString(strings.Repeat(" ", trackWidth-currentLen))
		}

		buf.WriteString("║\n")
	}

	return buf.String()
}

// findCarAt checks if a car should be rendered at this track position
// Maps the 2D car position to pseudo-3D track coordinates
func (r *Renderer) findCarAt(row, col, trackWidth, trackHeight int) string {
	// For simplicity, map Y position to row (with scaling)
	// and X position to column relative to track center

	for _, car := range r.state.Cars {
		// Scale car's Y position to track row (0-9)
		// Assuming car Y ranges from 0 to some max, map to track rows
		carRow := (car.Y * trackHeight) / 20 // Assuming Y range 0-20
		if carRow >= trackHeight {
			carRow = trackHeight - 1
		}

		// Map car X to track column
		depth := float64(carRow) / float64(trackHeight-1)
		leftMargin := int((1.0 - depth) * float64(trackWidth) / 4)
		roadWidth := trackWidth - 2*leftMargin

		if roadWidth > 2 {
			// Car X position relative to track center
			// Assuming X range 0-20, center at 10
			relativeX := car.X - 10
			carCol := leftMargin + 1 + (roadWidth-2)/2 + relativeX

			if carRow == row && carCol == col {
				return "█" // Car representation
			}
		}
	}

	return ""
}

// GetState returns the current renderer state (for testing)
func (r *Renderer) GetState() *State {
	return r.state
}
