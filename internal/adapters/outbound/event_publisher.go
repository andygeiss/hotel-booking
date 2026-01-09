package outbound

import (
	"context"
	"encoding/json"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/messaging"
)

// This file contains the implementation of the EventPublisher.
// It is defined in the domain/indexing package as an outbound port.
// It uses a messaging dispatcher from the cloud-native-utils package.

// EventPublisher represents an event publisher.
type EventPublisher struct {
	dispatcher messaging.Dispatcher
}

// NewEventPublisher creates a new event publisher.
func NewEventPublisher(dispatcher messaging.Dispatcher) *EventPublisher {
	return &EventPublisher{
		dispatcher: dispatcher,
	}
}

// Publish publishes an event.
func (ep *EventPublisher) Publish(ctx context.Context, e event.Event) error {
	// Encode the event to JSON.
	encoded, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// Create a new message with the encoded event.
	msg := messaging.NewMessage(e.Topic(), encoded)

	// Publish the message or return an error if it fails.
	if err := ep.dispatcher.Publish(ctx, msg); err != nil {
		return err
	}
	return nil
}
