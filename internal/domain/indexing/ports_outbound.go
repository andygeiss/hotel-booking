package indexing

import (
	"context"
	"go-ddd-hex-starter/internal/domain/event"

	"github.com/andygeiss/cloud-native-utils/resource"
)

// IndexRepository represents the repository for indexing.
// It provides methods for creating, retrieving, updating, deleting, and listing indexes.
// We will not reinvent the wheel and use the resource.Access type from the cloud-native-utils package.
type IndexRepository resource.Access[IndexID, Index]

// EventPublisher represents the publisher for events.
// It provides methods for publishing events.
// We will use the messaging.Dispatcher type from the cloud-native-utils package.
type EventPublisher interface {
	Publish(ctx context.Context, e event.Event) error
}
