package outbound_test

import (
	"context"
	"fmt"
	"go-ddd-hex-starter/internal/adapters/outbound"
	"go-ddd-hex-starter/internal/domain/indexing"
	"os"
	"testing"
)

// Every benchmark should follow the Benchmark_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the benchmarks to be easy to read and understand.

// We use the following benchmarks to create a baseline for our application.
// This will be used for Profile Guided Optimization (PGO) later.
// You can run these benchmarks with `just profile` and writes the results to `cpuprofile.pprof`.

// During the build process, we will use the results of these benchmarks to optimize our application.
// You can run the build process with `just build` and it will use the -pgo flag to optimize the application.

// We do not need to test the FileIndexRepository struct as it is a simple wrapper around the JsonFileAccess type.
// The JsonFileAccess type is already tested in the cloud-native-utils package.

const (
	BenchmarkMaxFileCount = 1000
	BenchmarkMaxFileSize  = 1024 * 1024 // 1 MB
)

func Benchmark_FileIndexRepository_Create_With_1000_Entries_Should_Be_Fast(b *testing.B) {
	// Arrange
	os.MkdirAll("testdata", 0755)
	defer os.RemoveAll("testdata")
	path := "testdata/index.json"
	repo := outbound.NewFileIndexRepository(path)

	// Create BenchmarkMaxFileCount number of indexing.FileInfo as a slice.
	// This will be used to create the index and benchmark the performance.
	files := make([]indexing.FileInfo, BenchmarkMaxFileCount)

	// Use range loop over int to initialize the slice
	for i := range files {
		files[i] = indexing.FileInfo{
			AbsPath: fmt.Sprintf("file%d.txt", i),
			Size:    BenchmarkMaxFileSize,
		}
	}

	// Create the Index instance by using the path as the ID.
	id := indexing.IndexID(path)
	index := indexing.NewIndex(id, files)
	ctx := context.Background()

	// Benchmark
	b.ResetTimer()
	for b.Loop() {
		repo.Create(ctx, id, index)
	}
}
