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
	fetcher := repofetcher.NewRepositoryContentFetcher(client)

	ctx := context.Background()
	results, err := fetcher.FetchAllRepositoriesContent(ctx, cfg, "README.md", "main")
	if err != nil {
		log.Fatalf("Error fetching repository content: %v", err)
	}
	for repo, content := range results {
		fmt.Printf("Repo: %s, Content: %d bytes\n", repo, len(content))
	}
}
