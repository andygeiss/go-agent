package outbound

import (
	"context"

	"github.com/andygeiss/go-agent/pkg/event"
)

// EventPublisher is a no-op implementation of the event.EventPublisher interface.
// It is used for demo purposes when event publishing is not needed.
type EventPublisher struct{}

// NewEventPublisher creates a new NoopPublisher instance.
func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

// Publish does nothing and returns nil.
func (p *EventPublisher) Publish(_ context.Context, _ event.Event) error {
	return nil
}
