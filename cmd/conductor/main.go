package main

import (
	"context"
	"fmt"
	"log"

	githubadapter "github.com/lerenn/conductor/pkg/adapters/github"
	"github.com/lerenn/conductor/pkg/config"
	"github.com/lerenn/conductor/pkg/repofetcher"
)

func main() {
	cfg, err := config.Load("configs")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	fmt.Printf("Loaded configuration: %+v\n", cfg)

	token := "" // TODO: load from env or config
	client := githubadapter.New(token)
	fetcher := repofetcher.NewRepositoryFilesFetcher(client)

	ctx := context.Background()
	if len(cfg.Repositories) == 0 {
		log.Fatalf("No repositories configured")
	}
	repo := cfg.Repositories[0]
	results, err := fetcher.FetchRepositoryFiles(ctx, repo.URL, "main", "README.md", "LICENSE")
	if err != nil {
		log.Fatalf("Error fetching repository files: %v", err)
	}
	for file, content := range results {
		fmt.Printf("File: %s, Content: %d bytes\n", file, len(content))
	}
}
