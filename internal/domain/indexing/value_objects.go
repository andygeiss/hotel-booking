package indexing

// EventFileIndexCreated represents a file index created event.
type EventFileIndexCreated struct {
	IndexID   IndexID `json:"index_id"`
	FileCount int     `json:"file_count"`
}

// NewEventFileIndexCreated creates a new EventFileIndexCreated instance.
func NewEventFileIndexCreated(id IndexID, count int) EventFileIndexCreated {
	return EventFileIndexCreated{
		IndexID:   id,
		FileCount: count,
	}
}

// Topic returns the topic for the event.
func (a EventFileIndexCreated) Topic() string {
	return "file_index_created"
}

// IndexID represents a value object and unique identifier for an index.
// It is used by the aggregate to identify the index.
type IndexID string
