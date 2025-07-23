package conductor

import (
	"context"
	"fmt"
	"log"

	"github.com/lerenn/conductor/pkg/adapters/github"
	"github.com/lerenn/conductor/pkg/config"
	"github.com/lerenn/conductor/pkg/repofetcher"
)

// Conductor represents the main conductor application that orchestrates
// repository file fetching and processing.
type Conductor struct {
	config  *config.Config
	client  github.Client
	fetcher repofetcher.Fetcher
}

// New creates a new Conductor instance with the given configuration and GitHub token.
func New(cfg *config.Config, token string) *Conductor {
	client := github.New(token)
	fetcher := repofetcher.New(client)
	return &Conductor{
		config:  cfg,
		client:  client,
		fetcher: fetcher,
	}
}

// Run executes the main conductor workflow, fetching files from configured repositories.
func (c *Conductor) Run(ctx context.Context) error {
	if len(c.config.Repositories) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	// For now, process the first repository
	// TODO: Process all repositories
	repo := c.config.Repositories[0]

	results, err := c.fetcher.FetchRepositoryFiles(ctx, repo.URL, "main", "README.md", "LICENSE")
	if err != nil {
		return fmt.Errorf("error fetching repository files: %w", err)
	}

	// Process and display results
	for file, content := range results {
		fmt.Printf("File: %s, Content: %d bytes\n", file, len(content))
	}

	return nil
}

// RunWithLogging executes the conductor workflow with logging.
func (c *Conductor) RunWithLogging(ctx context.Context) {
	fmt.Printf("Loaded configuration: %+v\n", c.config)

	if err := c.Run(ctx); err != nil {
		log.Fatalf("Error running conductor: %v", err)
	}
}
