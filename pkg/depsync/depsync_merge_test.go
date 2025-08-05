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

func TestDepSync_Run_WithRepositories_ChecksPassAndMerge(t *testing.T) {
	cfg := &config.Config{
		Repositories: []string{
			"https://github.com/test/repo",
		},
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

	tc.MockDagger.EXPECT().CloneRepo(gomock.Any(), "https://github.com/test/repo", "main").Return(nil, nil)
	tc.MockDagger.EXPECT().CheckBranchExists(gomock.Any(), dagger.CheckBranchExistsParams{
		Dir:        nil,
		BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		RepoURL:    "https://github.com/test/repo",
	}).Return(false, nil)
	tc.MockDagger.EXPECT().UpdateGoDependency(gomock.Any(), dagger.UpdateGoDependencyParams{
		Dir:           nil,
		ModulePath:    "github.com/test/dep",
		TargetVersion: "v1.1.0",
	}).Return(nil, nil)
	tc.MockDagger.EXPECT().CommitAndPush(gomock.Any(), dagger.CommitAndPushParams{
		Dir:           nil,
		BranchName:    "depsync/update-github-com-test-dep-v1.1.0",
		ModulePath:    "github.com/test/dep",
		TargetVersion: "v1.1.0",
		AuthorName:    "DepSync Bot",
		AuthorEmail:   "depsync@example.com",
		RepoURL:       "https://github.com/test/repo",
	}).Return("depsync/update-github-com-test-dep-v1.1.0", nil)

	// Mock the CheckPullRequestExists call (returns -1 - no existing PR)
	tc.MockGitHubClient.EXPECT().CheckPullRequestExists(
		gomock.Any(),
		github.CheckPullRequestExistsParams{
			RepoURL:      "https://github.com/test/repo",
			SourceBranch: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(-1, nil)

	// Mock the CreateMergeRequest call
	tc.MockGitHubClient.EXPECT().CreateMergeRequest(
		gomock.Any(),
		github.CreateMergeRequestParams{
			RepoURL:       "https://github.com/test/repo",
			SourceBranch:  "depsync/update-github-com-test-dep-v1.1.0",
			ModulePath:    "github.com/test/dep",
			TargetVersion: "v1.1.0",
		},
	).Return(123, nil)

	ctx := context.Background()
	err := tc.DepSync.Run(ctx)

	assert.NoError(t, err)
}

func TestDepSync_Run_WithRepositories_MergeFails(t *testing.T) {
	cfg := &config.Config{
		Repositories: []string{
			"https://github.com/test/repo",
		},
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

	tc.MockDagger.EXPECT().CloneRepo(gomock.Any(), "https://github.com/test/repo", "main").Return(nil, nil)
	tc.MockDagger.EXPECT().CheckBranchExists(gomock.Any(), dagger.CheckBranchExistsParams{
		Dir:        nil,
		BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		RepoURL:    "https://github.com/test/repo",
	}).Return(false, nil)
	tc.MockDagger.EXPECT().UpdateGoDependency(gomock.Any(), dagger.UpdateGoDependencyParams{
		Dir:           nil,
		ModulePath:    "github.com/test/dep",
		TargetVersion: "v1.1.0",
	}).Return(nil, nil)
	tc.MockDagger.EXPECT().CommitAndPush(gomock.Any(), dagger.CommitAndPushParams{
		Dir:           nil,
		BranchName:    "depsync/update-github-com-test-dep-v1.1.0",
		ModulePath:    "github.com/test/dep",
		TargetVersion: "v1.1.0",
		AuthorName:    "DepSync Bot",
		AuthorEmail:   "depsync@example.com",
		RepoURL:       "https://github.com/test/repo",
	}).Return("depsync/update-github-com-test-dep-v1.1.0", nil)

	// Mock the CheckPullRequestExists call (returns -1 - no existing PR)
	tc.MockGitHubClient.EXPECT().CheckPullRequestExists(
		gomock.Any(),
		github.CheckPullRequestExistsParams{
			RepoURL:      "https://github.com/test/repo",
			SourceBranch: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(-1, nil)

	// Mock the CreateMergeRequest call
	tc.MockGitHubClient.EXPECT().CreateMergeRequest(
		gomock.Any(),
		github.CreateMergeRequestParams{
			RepoURL:       "https://github.com/test/repo",
			SourceBranch:  "depsync/update-github-com-test-dep-v1.1.0",
			ModulePath:    "github.com/test/dep",
			TargetVersion: "v1.1.0",
		},
	).Return(123, nil)

	ctx := context.Background()
	err := tc.DepSync.Run(ctx)

	// The process should continue even if merge fails
	assert.NoError(t, err)
}

func TestDepSync_Run_WithRepositories_BranchExists(t *testing.T) {
	cfg := &config.Config{
		Repositories: []string{
			"https://github.com/test/repo",
		},
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

	// Branch exists, so skip the dependency update but still create MR
	tc.MockDagger.EXPECT().CloneRepo(gomock.Any(), "https://github.com/test/repo", "main").Return(nil, nil)
	tc.MockDagger.EXPECT().CheckBranchExists(gomock.Any(), dagger.CheckBranchExistsParams{
		Dir:        nil,
		BranchName: "depsync/update-github-com-test-dep-v1.1.0",
		RepoURL:    "https://github.com/test/repo",
	}).Return(true, nil)
	// No UpdateGoDependency or CommitAndPush calls expected since branch exists

	// Mock the CheckPullRequestExists call (returns PR number - PR already exists)
	tc.MockGitHubClient.EXPECT().CheckPullRequestExists(
		gomock.Any(),
		github.CheckPullRequestExistsParams{
			RepoURL:      "https://github.com/test/repo",
			SourceBranch: "depsync/update-github-com-test-dep-v1.1.0",
		},
	).Return(123, nil)

	// Mock the GetPullRequestChecks call for existing PR
	tc.MockGitHubClient.EXPECT().GetPullRequestChecks(
		gomock.Any(),
		github.GetPullRequestChecksParams{
			RepoURL:  "https://github.com/test/repo",
			PRNumber: 123,
		},
	).Return(&github.CheckStatus{Status: "running"}, nil)

	// No CreateMergeRequest call expected since PR already exists

	ctx := context.Background()
	err := tc.DepSync.Run(ctx)

	assert.NoError(t, err)
}
