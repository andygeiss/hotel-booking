package inbound_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

func Test_EventSubscriber_NewEventSubscriber_Should_Create_Instance(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}

	// Act
	subscriber := inbound.NewEventSubscriber(dispatcher)

	// Assert
	assert.That(t, "subscriber must not be nil", subscriber != nil, true)
}

func Test_EventSubscriber_Subscribe_With_Valid_Event_Should_Handle_Successfully(t *testing.T) {
	// Arrange
	handled := false
	var receivedEvent event.Event

	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)

	factory := func() event.Event {
		return indexing.NewEventFileIndexCreated()
	}
	handler := func(e event.Event) error {
		handled = true
		receivedEvent = e
		return nil
	}

	ctx := context.Background()

	// Act
	err := subscriber.Subscribe(ctx, indexing.EventTopicFileIndexCreated, factory, handler)

	// Simulate receiving a message
	eventData := indexing.NewEventFileIndexCreated().
		WithIndexID("test-id").
		WithFileCount(10)
	jsonData, errMarshal := json.Marshal(eventData)
	assert.That(t, "marshal err must be nil", errMarshal == nil, true)
	msg := messaging.Message{
		Topic: indexing.EventTopicFileIndexCreated,
		Data:  jsonData,
	}
	state, msgErr := dispatcher.fn(ctx, msg)

	// Assert
	assert.That(t, "subscribe err must be nil", err == nil, true)
	assert.That(t, "message err must be nil", msgErr == nil, true)
	assert.That(t, "state must be completed", state, messaging.MessageStateCompleted)
	assert.That(t, "handler must be called", handled, true)
	assert.That(t, "received event must not be nil", receivedEvent != nil, true)
}

func Test_EventSubscriber_Subscribe_With_Invalid_JSON_Should_Return_Error(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)

	factory := func() event.Event {
		return indexing.NewEventFileIndexCreated()
	}
	handler := func(e event.Event) error {
		return nil
	}

	ctx := context.Background()

	// Act
	_ = subscriber.Subscribe(ctx, indexing.EventTopicFileIndexCreated, factory, handler)

	// Simulate receiving an invalid message
	msg := messaging.Message{
		Topic: indexing.EventTopicFileIndexCreated,
		Data:  []byte("invalid json"),
	}
	state, msgErr := dispatcher.fn(ctx, msg)

	// Assert
	assert.That(t, "message err must not be nil", msgErr != nil, true)
	assert.That(t, "state must be failed", state, messaging.MessageStateFailed)
}

func Test_EventSubscriber_Subscribe_With_Handler_Error_Should_Return_Failed_State(t *testing.T) {
	// Arrange
	expectedErr := errors.New("handler error")
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)

	factory := func() event.Event {
		return indexing.NewEventFileIndexCreated()
	}
	handler := func(e event.Event) error {
		return expectedErr
	}

	ctx := context.Background()

	// Act
	_ = subscriber.Subscribe(ctx, indexing.EventTopicFileIndexCreated, factory, handler)

	// Simulate receiving a message
	eventData := indexing.NewEventFileIndexCreated()
	jsonData, errMarshal := json.Marshal(eventData)
	assert.That(t, "marshal err must be nil", errMarshal == nil, true)
	msg := messaging.Message{
		Topic: indexing.EventTopicFileIndexCreated,
		Data:  jsonData,
	}
	state, msgErr := dispatcher.fn(ctx, msg)

	// Assert
	assert.That(t, "message err must match", msgErr.Error(), expectedErr.Error())
	assert.That(t, "state must be failed", state, messaging.MessageStateFailed)
}

func Test_EventSubscriber_Subscribe_With_Correct_Topic_Should_Subscribe_To_Topic(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)
	expectedTopic := "custom.topic"

	factory := func() event.Event {
		return indexing.NewEventFileIndexCreated()
	}
	handler := func(e event.Event) error {
		return nil
	}

	ctx := context.Background()

	// Act
	err := subscriber.Subscribe(ctx, expectedTopic, factory, handler)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "subscribed topic must match", dispatcher.subscribedTopic, expectedTopic)
}

// mockDispatcher is a mock implementation of the messaging.Dispatcher interface.
type mockDispatcher struct {
	fn              service.Function[messaging.Message, messaging.MessageState]
	subscribedTopic string
}

// Publish is a mock implementation of the messaging.Dispatcher.Publish method.
func (m *mockDispatcher) Publish(ctx context.Context, message messaging.Message) error {
	_, err := m.fn(ctx, message)
	return err
}

// Subscribe is a mock implementation of the messaging.Dispatcher.Subscribe method.
func (m *mockDispatcher) Subscribe(ctx context.Context, topic string, fn service.Function[messaging.Message, messaging.MessageState]) error {
	m.fn = fn
	m.subscribedTopic = topic
	return nil
}
