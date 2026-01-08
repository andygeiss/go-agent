package tooling

import (
	"github.com/andygeiss/go-agent/pkg/agent"
)

// NewCalculateTool creates a new calculate tool that evaluates arithmetic expressions.
// Supports +, -, *, / operators with proper precedence and parentheses.
func NewCalculateTool() agent.Tool {
	return agent.Tool{
		ID: "calculate",
		Definition: agent.NewToolDefinition("calculate", "Perform a simple arithmetic calculation").
			WithParameter("expression", "The arithmetic expression to evaluate (e.g., '2 + 2')"),
		Func: Calculate,
	}
}

// NewGetCurrentTimeTool creates a new tool that returns the current date and time.
func NewGetCurrentTimeTool() agent.Tool {
	return agent.Tool{
		ID:         "get_current_time",
		Definition: agent.NewToolDefinition("get_current_time", "Get the current date and time"),
		Func:       GetCurrentTime,
	}
}
