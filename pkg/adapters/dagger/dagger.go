package dagger

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/cryptellation/depsync/pkg/logging"
	"go.uber.org/zap"
)

// UpdateGoDependencyParams contains parameters for UpdateGoDependency.
type UpdateGoDependencyParams struct {
	Dir           *dagger.Directory
	ModulePath    string
	TargetVersion string
}

// CheckBranchExistsParams contains parameters for CheckBranchExists.
type CheckBranchExistsParams struct {
	Dir        *dagger.Directory
	BranchName string
	RepoURL    string
}

// CommitAndPushParams contains parameters for CommitAndPush.
type CommitAndPushParams struct {
	Dir           *dagger.Directory
	BranchName    string
	ModulePath    string
	TargetVersion string
	AuthorName    string
	AuthorEmail   string
	RepoURL       string
}

// Dagger defines the interface for Dagger operations.
//
//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -destination=mock_dagger.gen.go -package=dagger . Dagger
type Dagger interface {
	CloneRepo(ctx context.Context, repoURL, branch string) (*dagger.Directory, error)
	UpdateGoDependency(ctx context.Context, params UpdateGoDependencyParams) (*dagger.Directory, error)
	CheckBranchExists(ctx context.Context, params CheckBranchExistsParams) (bool, error)
	CommitAndPush(ctx context.Context, params CommitAndPushParams) (string, error)
	Close() error
}

// daggerAdapter implements the Dagger interface.
type daggerAdapter struct {
	client      *dagger.Client
	githubToken string
}

// NewDagger returns a new instance implementing the Dagger interface.
func NewDagger(ctx context.Context, githubToken string) (Dagger, error) {
	client, err := dagger.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &daggerAdapter{
		client:      client,
		githubToken: githubToken,
	}, nil
}

// Close closes the Dagger client connection.
func (d *daggerAdapter) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// CloneRepo clones the given repo URL at the given branch using Dagger and returns the cloned directory.
func (d *daggerAdapter) CloneRepo(ctx context.Context, repoURL, branch string) (*dagger.Directory, error) {
	logger := logging.C(ctx)
	logger.Info("Cloning repository", zap.String("repo_url", repoURL), zap.String("branch", branch))

	// Set up the token as a Dagger secret
	secret := d.client.SetSecret("github_token", d.githubToken)

	// Use a container to perform the git clone
	container := d.client.Container().From("alpine/git").
		WithSecretVariable("GITHUB_TOKEN", secret).
		WithExec([]string{"sh", "-c",
			fmt.Sprintf(
				"git clone --depth=1 --branch %s https://$GITHUB_TOKEN@%s /repo", branch, repoURL[8:], // strip https://
			),
		})
	dir := container.Directory("/repo")

	// Check if the directory exists by listing files (fail fast)
	entries, err := dir.Entries(ctx)
	if err != nil {
		logger.Error("Failed to clone repository", zap.Error(err))
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}
	logger.Info("Repository cloned", zap.Strings("files", entries))
	return dir, nil
}

// UpdateGoDependency updates a Go dependency in the given directory to the specified version.
func (d *daggerAdapter) UpdateGoDependency(ctx context.Context, params UpdateGoDependencyParams) (
	*dagger.Directory, error) {
	logger := logging.C(ctx)
	logger.Info("Updating Go dependency",
		zap.String("module_path", params.ModulePath),
		zap.String("target_version", params.TargetVersion))

	// Use a Go container to perform the dependency update
	container := d.client.Container().From("golang:1.24-alpine").
		WithMountedDirectory("/repo", params.Dir).
		WithWorkdir("/repo").
		WithExec([]string{"go", "get", fmt.Sprintf("%s@%s", params.ModulePath, params.TargetVersion)})

	// Get the updated directory
	updatedDir := container.Directory("/repo")

	// Check if the update was successful by verifying the go.mod file exists
	entries, err := updatedDir.Entries(ctx)
	if err != nil {
		logger.Error("Failed to update dependency", zap.Error(err))
		return nil, fmt.Errorf("failed to update dependency: %w", err)
	}

	// Verify go.mod still exists
	if !contains(entries, "go.mod") {
		logger.Error("go.mod file not found after dependency update")
		return nil, fmt.Errorf("go.mod file not found after dependency update")
	}

	logger.Info("Dependency updated successfully",
		zap.String("module_path", params.ModulePath),
		zap.String("target_version", params.TargetVersion))
	return updatedDir, nil
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CheckBranchExists checks if a branch already exists in the remote repository.
func (d *daggerAdapter) CheckBranchExists(ctx context.Context, params CheckBranchExistsParams) (bool, error) {
	logger := logging.C(ctx)
	logger.Info("Checking if branch exists",
		zap.String("branch_name", params.BranchName),
		zap.String("repo_url", params.RepoURL))

	// Set up the token as a Dagger secret
	secret := d.client.SetSecret("github_token", d.githubToken)

	// Use a container to perform the git ls-remote operation
	container := d.client.Container().From("alpine/git").
		WithSecretVariable("GITHUB_TOKEN", secret).
		WithMountedDirectory("/repo", params.Dir).
		WithWorkdir("/repo")

	// Set up remote with authentication (same as in CommitAndPush)
	owner, repo := extractOwnerAndRepoFromURL(params.RepoURL)
	container = container.WithExec([]string{"sh", "-c",
		fmt.Sprintf("git remote set-url origin https://$GITHUB_TOKEN@github.com/%s/%s.git",
			owner, repo)})

	// Add cache-busting parameter to prevent Dagger from caching the git ls-remote result
	// This ensures we get fresh results each time, even if the operation signature is the same
	cacheBuster := fmt.Sprintf("check_%d", time.Now().UnixNano())
	container = container.WithEnvVariable("CACHE_BUSTER", cacheBuster)

	// Perform the git ls-remote operation
	lsRemoteOutput, err := container.WithExec([]string{"sh", "-c",
		fmt.Sprintf("git ls-remote --heads origin %s", params.BranchName)}).Stdout(ctx)
	if err != nil {
		logger.Error("Failed to check branch existence", zap.Error(err))
		return false, fmt.Errorf("failed to check branch existence: %w", err)
	}

	// Check if the output is empty (branch doesn't exist) or non-empty (branch exists)
	branchExists := strings.TrimSpace(lsRemoteOutput) != ""

	if branchExists {
		logger.Warn("Branch already exists, skipping dependency update",
			zap.String("branch_name", params.BranchName),
			zap.String("repo_url", params.RepoURL))
	} else {
		logger.Info("Branch does not exist, proceeding with dependency update",
			zap.String("branch_name", params.BranchName))
	}

	return branchExists, nil
}

// CommitAndPush commits the changes and pushes to a new branch.
func (d *daggerAdapter) CommitAndPush(ctx context.Context, params CommitAndPushParams) (string, error) {
	logger := logging.C(ctx)
	logger.Info("Committing and pushing changes",
		zap.String("module_path", params.ModulePath),
		zap.String("branch_name", params.BranchName))

	commitMessage := fmt.Sprintf("fix(dependencies): update %s to %s", params.ModulePath, params.TargetVersion)

	// Set up the token as a Dagger secret
	secret := d.client.SetSecret("github_token", d.githubToken)

	// Use a container to perform the git operations
	container := d.client.Container().From("alpine/git").
		WithSecretVariable("GITHUB_TOKEN", secret).
		WithMountedDirectory("/repo", params.Dir).
		WithWorkdir("/repo")

	// Add cache-busting parameter to prevent Dagger from caching the git operations
	// This ensures we get fresh results each time, even if the operation signature is the same
	cacheBuster := fmt.Sprintf("commit_%d", time.Now().UnixNano())
	container = container.WithEnvVariable("CACHE_BUSTER", cacheBuster)

	// Configure git user
	container = container.WithExec([]string{"git", "config", "user.name", params.AuthorName})
	container = container.WithExec([]string{"git", "config", "user.email", params.AuthorEmail})

	// Add and commit changes
	container = container.WithExec([]string{"git", "add", "."})
	container = container.WithExec([]string{"git", "commit", "-m", commitMessage})

	// Create and checkout new branch
	container = container.WithExec([]string{"git", "checkout", "-b", params.BranchName})

	// Set up remote with authentication and push
	owner, repo := extractOwnerAndRepoFromURL(params.RepoURL)
	container = container.WithExec([]string{"sh", "-c",
		fmt.Sprintf("git remote set-url origin https://$GITHUB_TOKEN@github.com/%s/%s.git",
			owner, repo)})

	// Push the branch
	_, err := container.WithExec([]string{"git", "push", "-u", "origin", params.BranchName}).Sync(ctx)
	if err != nil {
		logger.Error("Failed to push branch", zap.Error(err))
		return "", fmt.Errorf("failed to push branch: %w", err)
	}

	logger.Info("Successfully committed and pushed changes",
		zap.String("branch_name", params.BranchName),
		zap.String("commit_message", commitMessage))
	return params.BranchName, nil
}

// extractOwnerAndRepoFromURL extracts owner and repo from a GitHub URL like "https://github.com/owner/repo.git"
func extractOwnerAndRepoFromURL(repoURL string) (string, string) {
	// Remove https:// prefix and .git suffix
	cleanURL := strings.TrimPrefix(repoURL, "https://")
	cleanURL = strings.TrimSuffix(cleanURL, ".git")

	// Split by / and extract owner and repo
	parts := strings.Split(cleanURL, "/")
	if len(parts) >= 3 {
		return parts[1], parts[2]
	}
	return "", ""
}
