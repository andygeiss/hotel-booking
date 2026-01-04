package outbound

import (
	"go-ddd-hex-starter/internal/domain/indexing"

	"github.com/andygeiss/cloud-native-utils/resource"
)

// The Outbound package is responsible for interacting with external systems and services.
// It is responsible for writing data to files and reading data from files.
// It is also responsible for interacting with the database.
// It uses the context.Cancel() to cancel I/O operation and returns an error if the operation is canceled
// (either directly ctx.Err() or an I/O error that wraps it).

// We reuse the JsonFileAccess type from the cloud-native-utils package to handle JSON file access.
// This allows us to easily read and write JSON files without having to implement the logic ourselves.

// FileIndexRepository is a repository for indexing data stored in files.
type FileIndexRepository struct {
	resource.JsonFileAccess[indexing.IndexID, indexing.Index]
}

// NewFileIndexRepository creates a new FileIndexRepository instance.
func NewFileIndexRepository(filename string) indexing.IndexRepository {
	// We reuse the JsonFileAccess type from the cloud-native-utils package to handle JSON file access.
	return resource.NewJsonFileAccess[indexing.IndexID, indexing.Index](filename)
}
