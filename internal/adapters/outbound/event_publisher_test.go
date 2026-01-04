package outbound_test

import (
	"context"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
)

func Test_EventPublisher_With_An_Internal_Dispatcher_Should_Publish_Successfully(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{
		fn: func(ctx context.Context, msg messaging.Message) (messaging.MessageState, error) {
			return messaging.MessageState(0), nil
		},
	}
	publisher := outbound.NewEventPublisher(dispatcher)
	event := indexing.NewEventFileIndexCreated(indexing.IndexID("test-id"), 1)

	// Act
	err := publisher.Publish(context.Background(), event)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
}

// mockDispatcher is a mock implementation of the messaging.Dispatcher interface.
type mockDispatcher struct {
	fn service.Function[messaging.Message, messaging.MessageState]
}

// Publish is a mock implementation of the messaging.Dispatcher.Publish method.
func (a *mockDispatcher) Publish(ctx context.Context, message messaging.Message) error {
	_, err := a.fn(ctx, message)
	return err
}

// Subscribe is a mock implementation of the messaging.Dispatcher.Subscribe method.
func (a *mockDispatcher) Subscribe(ctx context.Context, topic string, fn service.Function[messaging.Message, messaging.MessageState]) error {
	a.fn = fn
	return nil
}
