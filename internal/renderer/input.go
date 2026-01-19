package renderer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Command represents a user input command
type Command struct {
	Type string
	Args []string
}

// InputHandler manages keyboard input and converts it to commands
// Commands are sent to the engine, NOT applied directly to state
type InputHandler struct {
	reader *bufio.Reader
}

// NewInputHandler creates a new input handler
func NewInputHandler() *InputHandler {
	return &InputHandler{
		reader: bufio.NewReader(os.Stdin),
	}
}

// ReadCommand reads and parses a command from stdin
// Returns nil if no command is available or on error
func (h *InputHandler) ReadCommand() (*Command, error) {
	fmt.Print("> ")
	line, err := h.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, nil
	}

	return &Command{
		Type: parts[0],
		Args: parts[1:],
	}, nil
}

// ParseCommand parses a command string for testing
func ParseCommand(input string) *Command {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return nil
	}

	return &Command{
		Type: parts[0],
		Args: parts[1:],
	}
}

// Available commands:
// - tick: Advance the simulation by one tick
// - ticks N: Advance by N ticks
// - quit: Exit the simulation
// - help: Show available commands
func (c *Command) Help() string {
	return `Available commands:
  tick       - Advance simulation by one tick
  ticks N    - Advance simulation by N ticks
  quit       - Exit the simulation
  help       - Show this help message
`
}

// IsQuit returns true if this is a quit command
func (c *Command) IsQuit() bool {
	return c.Type == "quit" || c.Type == "exit" || c.Type == "q"
}

// IsHelp returns true if this is a help command
func (c *Command) IsHelp() bool {
	return c.Type == "help" || c.Type == "h" || c.Type == "?"
}

// IsTick returns true if this is a tick command
func (c *Command) IsTick() bool {
	return c.Type == "tick" || c.Type == "t"
}

// IsTicks returns true if this is a ticks command
func (c *Command) IsTicks() bool {
	return c.Type == "ticks"
}
