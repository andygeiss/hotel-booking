package main

import (
	"context"
	"embed"
	"fmt"
	"go-ddd-hex-starter/internal/adapters/inbound"
	"go-ddd-hex-starter/internal/adapters/outbound"
	"go-ddd-hex-starter/internal/domain/indexing"
	"os"
)

// We use embed.FS to embed files into the binary.
// This allows us to include files such as templates, configuration files, etc., directly within the binary.
// This is useful for distributing applications with all necessary resources included.

//go:embed assets
var efs embed.FS

func main() {
	ctx := context.Background()
	in := inbound.NewFileReader()
	path := "./index.json"
	defer os.Remove(path)
	out := outbound.NewFileIndexRepository(path)

	fileInfos, err := in.ReadFileInfos(ctx, ".")
	if err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	indexID := indexing.IndexID(wd)
	index := indexing.NewIndex(indexID, fileInfos)

	out.Create(ctx, indexID, index)

	fmt.Printf("❯ main: index has %d files\n", len(fileInfos))
}
