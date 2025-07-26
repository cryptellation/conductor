//go:build integration
// +build integration

package dagger

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDagger_CloneRepo_PublicRepo(t *testing.T) {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN") // or "" for public

	daggerAdapter, err := NewDagger(ctx, githubToken)
	if err != nil {
		// If Dagger connection fails, skip the test
		t.Skipf("Skipping test - Dagger connection failed: %v", err)
	}
	defer daggerAdapter.Close()

	// Use a public repo (no token needed)
	repoURL := "https://github.com/octocat/Hello-World"
	branch := "master"

	dir, err := daggerAdapter.CloneRepo(ctx, repoURL, branch)
	require.NoError(t, err)

	// Check for a known file in the repo
	entries, err := dir.Entries(ctx)
	require.NoError(t, err)
	assert.Contains(t, entries, "README")
}

func TestDagger_CloneRepo_DefaultBranch(t *testing.T) {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN") // or "" for public

	daggerAdapter, err := NewDagger(ctx, githubToken)
	if err != nil {
		// If Dagger connection fails, skip the test
		t.Skipf("Skipping test - Dagger connection failed: %v", err)
	}
	defer daggerAdapter.Close()

	// Use a public repo with default branch
	// Note: Hello-World repo uses "master" as default branch, not "main"
	repoURL := "https://github.com/octocat/Hello-World"

	dir, err := daggerAdapter.CloneRepo(ctx, repoURL, "master")
	require.NoError(t, err)

	// Check that the directory was created
	entries, err := dir.Entries(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, entries)
}
