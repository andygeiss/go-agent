package outbound_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
)

func Test_EventPublisher_Publish_With_ValidEvent_Should_Succeed(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	evt := &mockEvent{EventTopic: "test.topic", Data: "test data"}

	// Act
	err := publisher.Publish(context.Background(), evt)

	// Assert
	assert.That(t, "must not return error", err, nil)
}

func Test_EventPublisher_Publish_With_ValidEvent_Should_PublishToDispatcher(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	evt := &mockEvent{EventTopic: "test.topic", Data: "test data"}

	// Act
	_ = publisher.Publish(context.Background(), evt)

	// Assert
	assert.That(t, "must publish one message", len(dispatcher.publishedMessages), 1)
}

func Test_EventPublisher_Publish_With_ValidEvent_Should_UseCorrectTopic(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	evt := &mockEvent{EventTopic: "agent.task_started", Data: "test"}

	// Act
	_ = publisher.Publish(context.Background(), evt)

	// Assert
	assert.That(t, "must use correct topic", dispatcher.publishedMessages[0].Topic, "agent.task_started")
}

func Test_EventPublisher_Publish_With_ValidEvent_Should_EncodePayloadAsJSON(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	evt := &mockEvent{EventTopic: "test.topic", Data: "test data"}

	// Act
	_ = publisher.Publish(context.Background(), evt)

	// Assert
	payload := string(dispatcher.publishedMessages[0].Data)
	assert.That(t, "must contain data in JSON", strings.Contains(payload, "test data"), true)
}

func Test_EventPublisher_Publish_With_DispatcherError_Should_ReturnError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("dispatcher error")
	dispatcher := &mockDispatcher{publishErr: expectedErr}
	publisher := outbound.NewEventPublisher(dispatcher)
	evt := &mockEvent{EventTopic: "test.topic", Data: "test"}

	// Act
	err := publisher.Publish(context.Background(), evt)

	// Assert
	assert.That(t, "must return dispatcher error", err, expectedErr)
}

func Test_EventPublisher_Publish_With_MultipleEvents_Should_PublishAll(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	events := []*mockEvent{
		{EventTopic: "event.one", Data: "first"},
		{EventTopic: "event.two", Data: "second"},
		{EventTopic: "event.three", Data: "third"},
	}

	// Act
	for _, evt := range events {
		_ = publisher.Publish(context.Background(), evt)
	}

	// Assert
	assert.That(t, "must publish all messages", len(dispatcher.publishedMessages), 3)
}

func Test_EventPublisher_Publish_With_MultipleEvents_Should_PreserveTopicOrder(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	_ = publisher.Publish(context.Background(), &mockEvent{EventTopic: "first", Data: "1"})
	_ = publisher.Publish(context.Background(), &mockEvent{EventTopic: "second", Data: "2"})

	// Assert
	assert.That(t, "first message topic", dispatcher.publishedMessages[0].Topic, "first")
	assert.That(t, "second message topic", dispatcher.publishedMessages[1].Topic, "second")
}

// mockEvent implements event.Event for testing.
type mockEvent struct {
	EventTopic string `json:"topic"`
	Data       string `json:"data"`
}

func (e *mockEvent) Topic() string {
	return e.EventTopic
}

// mockDispatcher implements messaging.Dispatcher for testing.
type mockDispatcher struct {
	publishErr        error
	publishedMessages []messaging.Message
}

func (d *mockDispatcher) Publish(_ context.Context, msg messaging.Message) error {
	if d.publishErr != nil {
		return d.publishErr
	}
	d.publishedMessages = append(d.publishedMessages, msg)
	return nil
}

func (d *mockDispatcher) Subscribe(_ context.Context, _ string, _ service.Function[messaging.Message, messaging.MessageState]) error {
	return nil
}
