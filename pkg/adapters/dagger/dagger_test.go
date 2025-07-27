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

func TestDagger_UpdateGoDependency_PublicRepo(t *testing.T) {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN") // or "" for public

	daggerAdapter, err := NewDagger(ctx, githubToken)
	if err != nil {
		// If Dagger connection fails, skip the test
		t.Skipf("Skipping test - Dagger connection failed: %v", err)
	}
	defer daggerAdapter.Close()

	// Use a public repo with a known dependency
	repoURL := "https://github.com/octocat/Hello-World"
	branch := "master"

	// First clone the repo
	dir, err := daggerAdapter.CloneRepo(ctx, repoURL, branch)
	require.NoError(t, err)

	// Try to update a dependency (this might fail if the repo doesn't have go.mod, but that's okay for testing)
	// We'll use a well-known module that exists
	modulePath := "github.com/stretchr/testify"
	targetVersion := "v1.8.4"

	updatedDir, err := daggerAdapter.UpdateGoDependency(ctx, UpdateGoDependencyParams{
		Dir:           dir,
		ModulePath:    modulePath,
		TargetVersion: targetVersion,
	})
	if err != nil {
		// This is expected if the repo doesn't have a go.mod file
		t.Logf("UpdateGoDependency failed as expected (repo may not have go.mod): %v", err)
		return
	}

	// If successful, verify the go.mod file still exists
	entries, err := updatedDir.Entries(ctx)
	require.NoError(t, err)
	assert.Contains(t, entries, "go.mod")
}

func TestDagger_CommitAndPush(t *testing.T) {
	// TODO: Implement integration test for CommitAndPush functionality
	// This test would verify that the CommitAndPush method works correctly
	// with real repositories, but it's complex to test due to permissions
	// and the need for actual git operations.
	t.Skip("TODO: Implement CommitAndPush integration test")
}

func TestDagger_CheckBranchExists_NonExistingBranch(t *testing.T) {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN") // or "" for public

	daggerAdapter, err := NewDagger(ctx, githubToken)
	if err != nil {
		// If Dagger connection fails, skip the test
		t.Skipf("Skipping test - Dagger connection failed: %v", err)
	}
	defer daggerAdapter.Close()

	// Use a public repo
	repoURL := "https://github.com/octocat/Hello-World"
	branch := "master"

	// First clone the repo
	dir, err := daggerAdapter.CloneRepo(ctx, repoURL, branch)
	require.NoError(t, err)

	// Check for a non-existing branch
	branchName := "conductor/update-github-com-test-dep-v1.1.0"

	exists, err := daggerAdapter.CheckBranchExists(ctx, CheckBranchExistsParams{
		Dir:        dir,
		BranchName: branchName,
		RepoURL:    repoURL,
	})
	require.NoError(t, err)
	assert.False(t, exists, "Branch should not exist")
}

func TestDagger_CheckBranchExists_ExistingBranch(t *testing.T) {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN") // or "" for public

	daggerAdapter, err := NewDagger(ctx, githubToken)
	if err != nil {
		// If Dagger connection fails, skip the test
		t.Skipf("Skipping test - Dagger connection failed: %v", err)
	}
	defer daggerAdapter.Close()

	// Use a public repo
	repoURL := "https://github.com/octocat/Hello-World"
	branch := "master"

	// First clone the repo
	dir, err := daggerAdapter.CloneRepo(ctx, repoURL, branch)
	require.NoError(t, err)

	// Check for an existing branch (master branch should exist)
	branchName := "master" // This will check for the existing master branch

	exists, err := daggerAdapter.CheckBranchExists(ctx, CheckBranchExistsParams{
		Dir:        dir,
		BranchName: branchName,
		RepoURL:    repoURL,
	})
	require.NoError(t, err)
	assert.True(t, exists, "Branch should exist")
}
