//go:build unit
// +build unit

package depsync

import (
	"context"
	"testing"

	"github.com/cryptellation/depsync/pkg/adapters/dagger"
	"github.com/cryptellation/depsync/pkg/adapters/github"
	"github.com/cryptellation/depsync/pkg/config"
	"github.com/cryptellation/depsync/pkg/depgraph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDepSync_Run_WithRepositories_DeleteConflictedPRsEnabled(t *testing.T) {
	cfg := &config.Config{
		Repositories: []string{
			"https://github.com/test/repo",
		},
		DeleteConflictedPRs: true,
		Git: config.GitConfig{
			Author: config.GitAuthor{
				Name:  "DepSync Bot",
				Email: "depsync@example.com",
			},
		},
	}

	tc := newTestDepSync(t, cfg)
	defer tc.MockController.Finish()
	defer tc.DepSync.Close()

	expectedResults := map[string][]byte{
		"go.mod": []byte("module github.com/test/repo\nrequire github.com/test/dep v1.0.0\n"),
	}

	tc.MockFetcher.EXPECT().
		Fetch(gomock.Any(), "https://github.com/test/repo", "main", "go.mod").
		Return(expectedResults, nil)

	mockGraph := map[string]*depgraph.Service{
		"github.com/test/repo": {
			ModulePath:   "github.com/test/repo",
			Dependencies: map[string]depgraph.Dependency{},
		},
	}
	tc.MockGraphBuilder.EXPECT().BuildGraph(gomock.Any()).Return(mockGraph, nil)

	tc.MockVersionDetector.EXPECT().DetectAndSetCurrentVersions(gomock.Any(), gomock.Any(), mockGraph).Return(nil)

	mismatches := map[string]map[string]depgraph.Mismatch{
		"github.com/test/repo": {
			"github.com/test/dep": {Actual: "v1.0.0", Latest: "v1.1.0"},
		},
	}
	tc.MockChecker.EXPECT().Check(mockGraph).Return(mismatches, nil)

	// Branch exists, so skip the dependency update
	tc.MockDagger.EXPECT().CloneRepo(gomock.Any(), "https://github.com/test/repo", "main").Return(nil, nil)
	tc.MockDagger.EXPECT().CheckBranchExists(gomock.Any(), dagger.CheckBranchExistsParams{
		Dir:        nil,
		BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		RepoURL:    "https://github.com/test/repo",
	}).Return(true, nil)

	// Mock the CheckPullRequestExists call (returns PR number - PR already exists)
	tc.MockGitHubClient.EXPECT().CheckPullRequestExists(
		gomock.Any(),
		github.CheckPullRequestExistsParams{
			RepoURL:      "https://github.com/test/repo",
			SourceBranch: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(123, nil)

	// Mock the CheckMergeConflicts call - conflicts detected
	tc.MockGitHubClient.EXPECT().CheckMergeConflicts(
		gomock.Any(),
		github.CheckMergeConflictsParams{
			RepoURL:  "https://github.com/test/repo",
			PRNumber: 123,
		},
	).Return(&github.MergeConflictInfo{
		HasConflicts:    true,
		ConflictedFiles: []string{"go.mod", "go.sum"},
	}, nil)

	// Mock the deletion operations
	tc.MockGitHubClient.EXPECT().DeletePullRequest(
		gomock.Any(),
		github.DeletePullRequestParams{
			RepoURL:  "https://github.com/test/repo",
			PRNumber: 123,
		},
	).Return(nil)

	tc.MockGitHubClient.EXPECT().DeleteBranch(
		gomock.Any(),
		github.DeleteBranchParams{
			RepoURL:    "https://github.com/test/repo",
			BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(nil)

	// Run the test
	err := tc.DepSync.Run(context.Background())
	assert.NoError(t, err)
}

func TestDepSync_Run_WithRepositories_DeleteConflictedPRsDisabled(t *testing.T) {
	cfg := &config.Config{
		Repositories: []string{
			"https://github.com/test/repo",
		},
		DeleteConflictedPRs: false,
		Git: config.GitConfig{
			Author: config.GitAuthor{
				Name:  "DepSync Bot",
				Email: "depsync@example.com",
			},
		},
	}

	tc := newTestDepSync(t, cfg)
	defer tc.MockController.Finish()
	defer tc.DepSync.Close()

	expectedResults := map[string][]byte{
		"go.mod": []byte("module github.com/test/repo\nrequire github.com/test/dep v1.0.0\n"),
	}

	tc.MockFetcher.EXPECT().
		Fetch(gomock.Any(), "https://github.com/test/repo", "main", "go.mod").
		Return(expectedResults, nil)

	mockGraph := map[string]*depgraph.Service{
		"github.com/test/repo": {
			ModulePath:   "github.com/test/repo",
			Dependencies: map[string]depgraph.Dependency{},
		},
	}
	tc.MockGraphBuilder.EXPECT().BuildGraph(gomock.Any()).Return(mockGraph, nil)

	tc.MockVersionDetector.EXPECT().DetectAndSetCurrentVersions(gomock.Any(), gomock.Any(), mockGraph).Return(nil)

	mismatches := map[string]map[string]depgraph.Mismatch{
		"github.com/test/repo": {
			"github.com/test/dep": {Actual: "v1.0.0", Latest: "v1.1.0"},
		},
	}
	tc.MockChecker.EXPECT().Check(mockGraph).Return(mismatches, nil)

	// Branch exists, so skip the dependency update
	tc.MockDagger.EXPECT().CloneRepo(gomock.Any(), "https://github.com/test/repo", "main").Return(nil, nil)
	tc.MockDagger.EXPECT().CheckBranchExists(gomock.Any(), dagger.CheckBranchExistsParams{
		Dir:        nil,
		BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		RepoURL:    "https://github.com/test/repo",
	}).Return(true, nil)

	// Mock the CheckPullRequestExists call (returns PR number - PR already exists)
	tc.MockGitHubClient.EXPECT().CheckPullRequestExists(
		gomock.Any(),
		github.CheckPullRequestExistsParams{
			RepoURL:      "https://github.com/test/repo",
			SourceBranch: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(123, nil)

	// Mock the check and merge operations (deletion is disabled, so continue with normal flow)
	tc.MockGitHubClient.EXPECT().GetPullRequestChecks(
		gomock.Any(),
		github.GetPullRequestChecksParams{
			RepoURL:  "https://github.com/test/repo",
			PRNumber: 123,
		},
	).Return(&github.CheckStatus{Status: "passed"}, nil)

	tc.MockGitHubClient.EXPECT().MergeMergeRequest(
		gomock.Any(),
		github.MergeMergeRequestParams{
			RepoURL:       "https://github.com/test/repo",
			PRNumber:      123,
			ModulePath:    "github.com/test/dep",
			TargetVersion: "v1.1.0",
		},
	).Return(nil)

	tc.MockGitHubClient.EXPECT().DeleteBranch(
		gomock.Any(),
		github.DeleteBranchParams{
			RepoURL:    "https://github.com/test/repo",
			BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(nil)

	// Run the test
	err := tc.DepSync.Run(context.Background())
	assert.NoError(t, err)
}

func TestDepSync_Run_WithRepositories_DeleteConflictedPRsError(t *testing.T) {
	cfg := &config.Config{
		Repositories: []string{
			"https://github.com/test/repo",
		},
		DeleteConflictedPRs: true,
		Git: config.GitConfig{
			Author: config.GitAuthor{
				Name:  "DepSync Bot",
				Email: "depsync@example.com",
			},
		},
	}

	tc := newTestDepSync(t, cfg)
	defer tc.MockController.Finish()
	defer tc.DepSync.Close()

	expectedResults := map[string][]byte{
		"go.mod": []byte("module github.com/test/repo\nrequire github.com/test/dep v1.0.0\n"),
	}

	tc.MockFetcher.EXPECT().
		Fetch(gomock.Any(), "https://github.com/test/repo", "main", "go.mod").
		Return(expectedResults, nil)

	mockGraph := map[string]*depgraph.Service{
		"github.com/test/repo": {
			ModulePath:   "github.com/test/repo",
			Dependencies: map[string]depgraph.Dependency{},
		},
	}
	tc.MockGraphBuilder.EXPECT().BuildGraph(gomock.Any()).Return(mockGraph, nil)

	tc.MockVersionDetector.EXPECT().DetectAndSetCurrentVersions(gomock.Any(), gomock.Any(), mockGraph).Return(nil)

	mismatches := map[string]map[string]depgraph.Mismatch{
		"github.com/test/repo": {
			"github.com/test/dep": {Actual: "v1.0.0", Latest: "v1.1.0"},
		},
	}
	tc.MockChecker.EXPECT().Check(mockGraph).Return(mismatches, nil)

	// Branch exists, so skip the dependency update
	tc.MockDagger.EXPECT().CloneRepo(gomock.Any(), "https://github.com/test/repo", "main").Return(nil, nil)
	tc.MockDagger.EXPECT().CheckBranchExists(gomock.Any(), dagger.CheckBranchExistsParams{
		Dir:        nil,
		BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		RepoURL:    "https://github.com/test/repo",
	}).Return(true, nil)

	// Mock the CheckPullRequestExists call (returns PR number - PR already exists)
	tc.MockGitHubClient.EXPECT().CheckPullRequestExists(
		gomock.Any(),
		github.CheckPullRequestExistsParams{
			RepoURL:      "https://github.com/test/repo",
			SourceBranch: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(123, nil)

	// Mock the CheckMergeConflicts call - conflicts detected
	tc.MockGitHubClient.EXPECT().CheckMergeConflicts(
		gomock.Any(),
		github.CheckMergeConflictsParams{
			RepoURL:  "https://github.com/test/repo",
			PRNumber: 123,
		},
	).Return(&github.MergeConflictInfo{
		HasConflicts:    true,
		ConflictedFiles: []string{"go.mod", "go.sum"},
	}, nil)

	// Mock the deletion operations - PR deletion fails
	tc.MockGitHubClient.EXPECT().DeletePullRequest(
		gomock.Any(),
		github.DeletePullRequestParams{
			RepoURL:  "https://github.com/test/repo",
			PRNumber: 123,
		},
	).Return(assert.AnError)

	// Run the test - should return error
	err := tc.DepSync.Run(context.Background())
	assert.Error(t, err)
}
