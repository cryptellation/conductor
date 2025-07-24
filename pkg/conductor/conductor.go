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

	for _, repo := range c.config.Repositories {
		fmt.Printf("Fetching go.mod for repository: %s (%s)\n", repo.Name, repo.URL)
		results, err := c.fetcher.FetchRepositoryFiles(ctx, repo.URL, "main", "go.mod")
		if err != nil {
			return fmt.Errorf("error fetching go.mod for %s: %w", repo.Name, err)
		}
		content, ok := results["go.mod"]
		if !ok {
			fmt.Printf("go.mod not found in repository: %s\n", repo.Name)
			continue
		}
		fmt.Printf("Repository: %s, go.mod size: %d bytes\n", repo.Name, len(content))
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
