package inbound

import (
	"context"
	"encoding/json"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
)

// This file contains the implementation of the EventSubscriber.
// It is defined in the domain/event package as an inbound port.
// It uses a messaging dispatcher from the cloud-native-utils package.

// EventSubscriber represents an event subscriber.
type EventSubscriber struct {
	dispatcher messaging.Dispatcher
}

// NewEventSubscriber creates a new event subscriber.
func NewEventSubscriber(dispatcher messaging.Dispatcher) *EventSubscriber {
	return &EventSubscriber{
		dispatcher: dispatcher,
	}
}

// Subscribe subscribes to a topic and calls the provided function when an event is received.
// The factory function creates a new instance of the concrete event type for unmarshaling.
// The handler can then type-assert the event to access its fields.
func (es *EventSubscriber) Subscribe(ctx context.Context, topic string, factory func() event.Event, handler func(e event.Event) error) error {
	// Wrap the domain event handler into a messaging function.
	messageFn := func(msg messaging.Message) (messaging.MessageState, error) {
		// Create a new instance of the event type using the factory.
		evt := factory()

		// Decode the message payload into the concrete event type.
		if err := json.Unmarshal(msg.Data, evt); err != nil {
			return messaging.MessageStateFailed, err
		}

		// Call the provided domain event handler.
		if err := handler(evt); err != nil {
			return messaging.MessageStateFailed, err
		}
		return messaging.MessageStateCompleted, nil
	}

	// Subscribe to the topic using the dispatcher.
	return es.dispatcher.Subscribe(ctx, topic, service.Wrap(messageFn))
}
