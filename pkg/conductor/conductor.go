package conductor

import (
	"context"
	"fmt"
	"strings"

	"github.com/cryptellation/depsync/pkg/adapters/dagger"
	"github.com/cryptellation/depsync/pkg/adapters/github"
	"github.com/cryptellation/depsync/pkg/config"
	"github.com/cryptellation/depsync/pkg/depgraph"
	"github.com/cryptellation/depsync/pkg/logging"
	"github.com/cryptellation/depsync/pkg/repo"
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
	dagger          dagger.Dagger
}

// New creates a new Conductor instance with the given configuration and GitHub token.
func New(cfg *config.Config, token string) (*Conductor, error) {
	client := github.New(token)

	// Create dagger adapter with context
	ctx := context.Background()
	daggerAdapter, err := dagger.NewDagger(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create dagger adapter: %w", err)
	}

	return &Conductor{
		config:          cfg,
		client:          client,
		fetcher:         repo.NewFilesFetcher(client),
		graphBuilder:    depgraph.NewGraphBuilder(),
		versionDetector: repo.NewVersionDetector(),
		checker:         depgraph.NewInconsistencyChecker(),
		dagger:          daggerAdapter,
	}, nil
}

// Close closes the Conductor and its resources.
func (c *Conductor) Close() error {
	if c.dagger != nil {
		return c.dagger.Close()
	}
	return nil
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
	if len(mismatches) == 0 {
		return nil
	}
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
	// Call the fixModules method to handle dependency updates
	if err := c.fixModules(ctx, mismatches); err != nil {
		return fmt.Errorf("failed to fix modules: %w", err)
	}

	return nil
}

// fixModules handles the dependency update workflow using the Dagger adapter.
func (c *Conductor) fixModules(ctx context.Context, mismatches map[string]map[string]depgraph.Mismatch) error {
	logger := logging.C(ctx)
	logger.Info("Starting fixModules workflow", zap.Int("service_count", len(mismatches)))

	// Iterate mismatches and clone each repo for each dependency update
	for service, deps := range mismatches {
		logger.Info("Processing service", zap.String("service", service))

		// Convert Go module path to GitHub URL
		// Format: github.com/x/y -> https://github.com/x/y
		repoURL := "https://" + service

		// Update each dependency for this service
		for dep, mismatch := range deps {
			branchName, err := c.updateDependency(ctx, service, dep, mismatch, repoURL)
			if err != nil {
				return err
			}

			// Always attempt MR creation, even if branch already existed
			// In the future, we will detect if the MR already exists
			if err := c.manageMergeRequest(ctx, service, dep, mismatch, repoURL, branchName); err != nil {
				return err
			}
		}

		logger.Info("All dependencies processed for service",
			zap.String("service", service),
			zap.String("repo_url", repoURL))
	}

	logger.Info("fixModules workflow completed successfully")
	return nil
}

// updateDependency updates a single dependency for a service.
//
//nolint:funlen // This function orchestrates a complex workflow that's difficult to break down further
func (c *Conductor) updateDependency(ctx context.Context, service, dep string, mismatch depgraph.Mismatch,
	repoURL string) (string, error) {
	logger := logging.C(ctx)
	logger.Info("Updating dependency",
		zap.String("service", service),
		zap.String("dependency", dep),
		zap.String("from", mismatch.Actual),
		zap.String("to", mismatch.Latest))

	// Clone the repo fresh for each dependency update
	dir, err := c.dagger.CloneRepo(ctx, repoURL, "main")
	if err != nil {
		logger.Error("Failed to clone repo for service", zap.String("service", service), zap.Error(err))
		return "", err
	}

	// Generate branch name
	branchName := generateBranchName(dep, mismatch.Latest)

	// Check if the branch already exists
	branchExists, err := c.dagger.CheckBranchExists(ctx, dagger.CheckBranchExistsParams{
		Dir:        dir,
		BranchName: branchName,
		RepoURL:    repoURL,
	})
	if err != nil {
		logger.Error("Failed to check branch existence",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.Error(err))
		return "", err
	}

	// If branch exists, skip dependency update but return the branch name
	if branchExists {
		logger.Warn("Branch already exists, skipping dependency update",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.String("target_version", mismatch.Latest),
			zap.String("branch_name", branchName))
		return branchName, nil
	}

	// Update the dependency
	updatedDir, err := c.dagger.UpdateGoDependency(ctx, dagger.UpdateGoDependencyParams{
		Dir:           dir,
		ModulePath:    dep,
		TargetVersion: mismatch.Latest,
	})
	if err != nil {
		logger.Error("Failed to update dependency",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.Error(err))
		return "", err
	}

	logger.Info("Dependency updated successfully",
		zap.String("service", service),
		zap.String("dependency", dep),
		zap.String("repo_url", repoURL))

	// Commit and push the changes
	_, err = c.dagger.CommitAndPush(ctx, dagger.CommitAndPushParams{
		Dir:         updatedDir,
		BranchName:  branchName,
		ModulePath:  dep,
		AuthorName:  c.config.Git.Author.Name,
		AuthorEmail: c.config.Git.Author.Email,
		RepoURL:     repoURL,
	})
	if err != nil {
		logger.Error("Failed to commit and push changes",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.Error(err))
		return "", err
	}

	logger.Info("Successfully committed and pushed changes",
		zap.String("service", service),
		zap.String("dependency", dep),
		zap.String("branch_name", branchName),
		zap.String("repo_url", repoURL))

	return branchName, nil
}

// manageMergeRequest creates a merge request for the updated dependency.
func (c *Conductor) manageMergeRequest(ctx context.Context, service, dep string, mismatch depgraph.Mismatch,
	repoURL, branchName string) error {
	logger := logging.C(ctx)
	logger.Info("Creating merge request",
		zap.String("service", service),
		zap.String("dependency", dep),
		zap.String("from", mismatch.Actual),
		zap.String("to", mismatch.Latest))

	// Check if a pull request already exists for this branch
	prNumber, err := c.checkExistingPullRequest(ctx, service, dep, repoURL, branchName)
	if err != nil {
		return err
	}

	// If no PR exists, create it
	if prNumber == -1 {
		prNumber, err = c.createMergeRequest(ctx, service, dep, mismatch, repoURL, branchName)
		if err != nil {
			return err
		}
	}

	// Check and log CI/CD status
	c.checkAndLogCIStatus(ctx, service, dep, repoURL, prNumber)

	return nil
}

// checkExistingPullRequest checks if a pull request already exists for the given branch.
func (c *Conductor) checkExistingPullRequest(ctx context.Context, service, dep, repoURL, branchName string) (
	int, error) {
	logger := logging.C(ctx)
	prNumber, err := c.client.CheckPullRequestExists(ctx, github.CheckPullRequestExistsParams{
		RepoURL:      repoURL,
		SourceBranch: branchName,
	})
	if err != nil {
		logger.Error("Failed to check if pull request exists",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.String("branch_name", branchName),
			zap.Error(err))
		return -1, err
	}

	if prNumber != -1 {
		logger.Warn("Pull request already exists, skipping creation",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.String("branch_name", branchName),
			zap.String("repo_url", repoURL),
			zap.Int("pr_number", prNumber))
	}

	return prNumber, nil
}

// createMergeRequest creates a new merge request.
func (c *Conductor) createMergeRequest(ctx context.Context, service, dep string, mismatch depgraph.Mismatch,
	repoURL, branchName string) (int, error) {
	logger := logging.C(ctx)
	prNumber, err := c.client.CreateMergeRequest(ctx, github.CreateMergeRequestParams{
		RepoURL:       repoURL,
		SourceBranch:  branchName,
		ModulePath:    dep,
		TargetVersion: mismatch.Latest,
	})
	if err != nil {
		logger.Error("Failed to create merge request",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.String("branch_name", branchName),
			zap.Error(err))
		return -1, err
	}

	logger.Info("Successfully created merge request",
		zap.String("service", service),
		zap.String("dependency", dep),
		zap.String("branch_name", branchName),
		zap.String("repo_url", repoURL),
		zap.Int("pr_number", prNumber))

	return prNumber, nil
}

// checkAndLogCIStatus checks the CI/CD status and logs the result.
func (c *Conductor) checkAndLogCIStatus(ctx context.Context, service, dep, repoURL string, prNumber int) {
	logger := logging.C(ctx)
	checkStatus, err := c.client.GetPullRequestChecks(ctx, github.GetPullRequestChecksParams{
		RepoURL:  repoURL,
		PRNumber: prNumber,
	})
	if err != nil {
		logger.Error("Failed to get pull request checks",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.Int("pr_number", prNumber),
			zap.Error(err))
		// Continue with other MRs, don't fail the entire process
		return
	}

	// Log the check status
	switch checkStatus.Status {
	case "running":
		logger.Info("CI/CD checks are still running",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.Int("pr_number", prNumber))
	case "passed":
		logger.Info("CI/CD checks have passed",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.Int("pr_number", prNumber))
	case "failed":
		logger.Warn("CI/CD checks have failed - manual intervention required",
			zap.String("service", service),
			zap.String("dependency", dep),
			zap.Int("pr_number", prNumber))
	}
}

// sanitizeBranchName sanitizes a string to be used as a git branch name.
func sanitizeBranchName(name string) string {
	// Replace invalid characters with hyphens
	invalidChars := []string{"/", ".", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "-")
	}
	// Remove consecutive hyphens
	result = strings.ReplaceAll(result, "--", "-")
	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")
	return result
}

// generateBranchName generates a consistent branch name for dependency updates.
func generateBranchName(modulePath, targetVersion string) string {
	return fmt.Sprintf("conductor/update-%s-%s", sanitizeBranchName(modulePath), targetVersion)
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
