package dagger

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/lerenn/conductor/pkg/logging"
	"go.uber.org/zap"
)

// Dagger defines the interface for Dagger operations.
//
//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -destination=mock_dagger.gen.go -package=dagger . Dagger
type Dagger interface {
	CloneRepo(ctx context.Context, repoURL, branch string) (*dagger.Directory, error)
	UpdateGoDependency(ctx context.Context, dir *dagger.Directory, modulePath,
		targetVersion string) (*dagger.Directory, error)
	CommitAndPush(ctx context.Context, dir *dagger.Directory, modulePath, targetVersion,
		authorName, authorEmail, repoURL string) (string, error)
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
func (d *daggerAdapter) UpdateGoDependency(ctx context.Context, dir *dagger.Directory,
	modulePath, targetVersion string) (*dagger.Directory, error) {
	logger := logging.C(ctx)
	logger.Info("Updating Go dependency",
		zap.String("module_path", modulePath),
		zap.String("target_version", targetVersion))

	// Use a Go container to perform the dependency update
	container := d.client.Container().From("golang:1.24-alpine").
		WithMountedDirectory("/repo", dir).
		WithWorkdir("/repo").
		WithExec([]string{"go", "get", fmt.Sprintf("%s@%s", modulePath, targetVersion)})

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
		zap.String("module_path", modulePath),
		zap.String("target_version", targetVersion))
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

// CommitAndPush commits the changes and pushes to a new branch.
func (d *daggerAdapter) CommitAndPush(ctx context.Context, dir *dagger.Directory, modulePath, targetVersion,
	authorName, authorEmail, repoURL string) (string, error) {
	logger := logging.C(ctx)
	logger.Info("Committing and pushing changes",
		zap.String("module_path", modulePath),
		zap.String("target_version", targetVersion))

	// Generate branch name
	branchName := fmt.Sprintf("conductor/update-%s-%s", sanitizeBranchName(modulePath), targetVersion)
	commitMessage := fmt.Sprintf("fix(dependencies): update %s to %s", modulePath, targetVersion)

	// Set up the token as a Dagger secret
	secret := d.client.SetSecret("github_token", d.githubToken)

	// Use a container to perform the git operations
	container := d.client.Container().From("alpine/git").
		WithSecretVariable("GITHUB_TOKEN", secret).
		WithMountedDirectory("/repo", dir).
		WithWorkdir("/repo").
		WithExec([]string{"git", "config", "user.name", authorName}).
		WithExec([]string{"git", "config", "user.email", authorEmail}).
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", commitMessage}).
		WithExec([]string{"git", "checkout", "-b", branchName})

	// Set up remote with authentication and push
	owner, repo := extractOwnerAndRepoFromURL(repoURL)
	container = container.WithExec([]string{"sh", "-c",
		fmt.Sprintf("git remote set-url origin https://$GITHUB_TOKEN@github.com/%s/%s.git",
			owner, repo)}).
		WithExec([]string{"git", "push", "-u", "origin", branchName})

	// Execute the container to perform the operations
	_, err := container.Sync(ctx)
	if err != nil {
		logger.Error("Failed to commit and push changes", zap.Error(err))
		return "", fmt.Errorf("failed to commit and push changes: %w", err)
	}

	logger.Info("Successfully committed and pushed changes",
		zap.String("branch_name", branchName),
		zap.String("commit_message", commitMessage))
	return branchName, nil
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
