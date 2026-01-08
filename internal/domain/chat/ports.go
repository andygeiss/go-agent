package chat

import (
	"context"

	"github.com/andygeiss/go-agent/pkg/agent"
)

// TaskRunner executes tasks for an agent.
type TaskRunner interface {
	// RunTask executes a task and returns the result.
	RunTask(ctx context.Context, agent *agent.Agent, task *agent.Task) (agent.Result, error)
}
