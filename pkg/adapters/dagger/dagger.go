package dagger

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/lerenn/conductor/pkg/logging"
	"go.uber.org/zap"
)

// Dagger defines the interface for Dagger operations.
//
//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -destination=mock_dagger.gen.go -package=dagger . Dagger
type Dagger interface {
	CloneRepo(ctx context.Context, repoURL, branch string) (*dagger.Directory, error)
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
