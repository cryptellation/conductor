package conductor

import (
	"context"
	"fmt"

	"github.com/lerenn/conductor/pkg/adapters/github"
	"github.com/lerenn/conductor/pkg/config"
	"github.com/lerenn/conductor/pkg/depgraph"
	"github.com/lerenn/conductor/pkg/logging"
	"github.com/lerenn/conductor/pkg/repo"
	"go.uber.org/zap"
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

	c.printDependencyGraph(ctx, graph)
	c.printCurrentVersions(ctx, graph)

	mismatches, err := c.checker.Check(graph)
	if err != nil {
		return fmt.Errorf("failed to check for inconsistencies: %w", err)
	}
	if len(mismatches) > 0 {
		logging.C(ctx).Warn("Version inconsistencies detected")
		for svc, deps := range mismatches {
			for dep, mismatch := range deps {
				logging.C(ctx).Warn("Dependency version mismatch",
					zap.String("service", svc),
					zap.String("dependency", dep),
					zap.String("actual", mismatch.Actual),
					zap.String("latest", mismatch.Latest),
				)
			}
		}
	}

	return nil
}

// fetchModules fetches go.mod files and builds the input map for the dependency graph builder.
func (c *Conductor) fetchModules(ctx context.Context) (map[string]depgraph.RepoModule, error) {
	modules := make(map[string]depgraph.RepoModule)
	for _, repo := range c.config.Repositories {
		logging.C(ctx).Info("Fetching go.mod for repository",
			zap.String("name", repo.Name),
			zap.String("url", repo.URL),
		)
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
		logging.C(ctx).Info("Repository module info",
			zap.String("name", repo.Name),
			zap.String("module_path", modulePath),
			zap.Int("go_mod_size", len(content)),
		)
	}
	return modules, nil
}

// printDependencyGraph prints the dependency graph in a readable format.
func (c *Conductor) printDependencyGraph(ctx context.Context, graph map[string]*depgraph.Service) {
	logging.C(ctx).Info("Dependency graph:")
	for module, svc := range graph {
		logging.C(ctx).Info("Module dependencies",
			zap.String("module", module),
			zap.Strings("dependencies", depKeys(svc.Dependencies)),
		)
	}
}

func depKeys(m map[string]depgraph.Dependency) []string {
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}

// printCurrentVersions prints the module path and CurrentVersion for each root service.
func (c *Conductor) printCurrentVersions(ctx context.Context, graph map[string]*depgraph.Service) {
	logging.C(ctx).Info("Detected versions:")
	for module, svc := range graph {
		logging.C(ctx).Info("Module version",
			zap.String("module", module),
			zap.String("latest_version", svc.LatestVersion),
		)
	}
}

// RunWithLogging executes the conductor workflow with logging.
func (c *Conductor) RunWithLogging(ctx context.Context) {
	logging.C(ctx).Info("Loaded configuration", zap.Any("config", c.config))

	if err := c.Run(ctx); err != nil {
		logging.C(ctx).Fatal("Error running conductor", zap.Error(err))
	}
}
