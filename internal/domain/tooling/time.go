package tooling

import (
	"context"
	"time"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// NewGetCurrentTimeTool creates a new tool that returns the current date and time.
func NewGetCurrentTimeTool() agent.Tool {
	return agent.Tool{
		ID:         "get_current_time",
		Definition: agent.NewToolDefinition("get_current_time", "Get the current date and time"),
		Func:       GetCurrentTime,
	}
}

// GetCurrentTime returns the current date and time in RFC3339 format.
func GetCurrentTime(_ context.Context, _ string) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}
