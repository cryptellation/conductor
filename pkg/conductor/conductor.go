package conductor

import (
	"context"
	"fmt"
	"log"

	"github.com/lerenn/conductor/pkg/adapters/github"
	"github.com/lerenn/conductor/pkg/config"
	"github.com/lerenn/conductor/pkg/depgraph"
	"github.com/lerenn/conductor/pkg/repo"
	"golang.org/x/mod/modfile"
)

// Conductor represents the main conductor application that orchestrates
// repository file fetching and processing.
type Conductor struct {
	config          *config.Config
	client          github.Client
	fetcher         repo.FilesFetcher
	graphBuilder    depgraph.GraphBuilder
	versionDetector repo.VersionDetector
	checker         depgraph.InconsistencyChecker
}

// New creates a new Conductor instance with the given configuration and GitHub token.
func New(cfg *config.Config, token string) *Conductor {
	client := github.New(token)
	return &Conductor{
		config:          cfg,
		client:          client,
		fetcher:         repo.NewFilesFetcher(client),
		graphBuilder:    depgraph.NewGraphBuilder(),
		versionDetector: repo.NewVersionDetector(),
		checker:         depgraph.NewInconsistencyChecker(),
	}
}

// Run executes the main conductor workflow, fetching files from configured repositories.
func (c *Conductor) Run(ctx context.Context) error {
	if len(c.config.Repositories) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	modules, err := c.fetchModules(ctx)
	if err != nil {
		return err
	}

	graph, err := c.graphBuilder.BuildGraph(modules)
	if err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	err = c.versionDetector.DetectAndSetCurrentVersions(ctx, c.client, graph)
	if err != nil {
		return fmt.Errorf("failed to detect versions: %w", err)
	}

	c.printDependencyGraph(graph)
	c.printCurrentVersions(graph)

	mismatches, err := c.checker.Check(graph)
	if err != nil {
		return fmt.Errorf("failed to check for inconsistencies: %w", err)
	}
	if len(mismatches) > 0 {
		fmt.Println("Version inconsistencies detected:")
		for svc, deps := range mismatches {
			for dep, mismatch := range deps {
				fmt.Printf("- %s depends on %s: actual=%s, latest=%s\n", svc, dep, mismatch.Actual, mismatch.Latest)
			}
		}
	}

	return nil
}

// fetchModules fetches go.mod files and builds the input map for the dependency graph builder.
func (c *Conductor) fetchModules(ctx context.Context) (map[string]depgraph.RepoModule, error) {
	modules := make(map[string]depgraph.RepoModule)
	for _, repo := range c.config.Repositories {
		fmt.Printf("Fetching go.mod for repository: %s (%s)\n", repo.Name, repo.URL)
		results, err := c.fetcher.Fetch(ctx, repo.URL, "main", "go.mod")
		if err != nil {
			return nil, fmt.Errorf("error fetching go.mod for %s: %w", repo.Name, err)
		}
		content, ok := results["go.mod"]
		if !ok {
			return nil, fmt.Errorf("go.mod not found in repository: %s", repo.Name)
		}
		mf, err := modfile.Parse("go.mod", content, nil)
		if err != nil || mf.Module == nil {
			return nil, fmt.Errorf("could not parse module path for repo %s: %w", repo.Name, err)
		}
		modulePath := mf.Module.Mod.Path
		modules[modulePath] = depgraph.RepoModule{
			RepoURL:      repo.URL,
			GoModContent: content,
		}
		fmt.Printf("Repository: %s, module path: %s, go.mod size: %d bytes\n", repo.Name, modulePath, len(content))
	}
	return modules, nil
}

// printDependencyGraph prints the dependency graph in a readable format.
func (c *Conductor) printDependencyGraph(graph map[string]*depgraph.Service) {
	fmt.Println("Dependency graph:")
	for module, svc := range graph {
		fmt.Printf("- %s:\n", module)
		for dep := range svc.Dependencies {
			fmt.Printf("    depends on: %s\n", dep)
		}
	}
}

// printCurrentVersions prints the module path and CurrentVersion for each root service.
func (c *Conductor) printCurrentVersions(graph map[string]*depgraph.Service) {
	fmt.Println("Detected versions:")
	for module, svc := range graph {
		fmt.Printf("- %s: %s\n", module, svc.LatestVersion)
	}
}

// RunWithLogging executes the conductor workflow with logging.
func (c *Conductor) RunWithLogging(ctx context.Context) {
	fmt.Printf("Loaded configuration: %+v\n", c.config)

	if err := c.Run(ctx); err != nil {
		log.Fatalf("Error running conductor: %v", err)
	}
}
