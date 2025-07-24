package conductor

import (
	"context"
	"testing"

	"github.com/lerenn/conductor/pkg/config"
	"github.com/lerenn/conductor/pkg/depgraph"
	"github.com/lerenn/conductor/pkg/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConductor_New(t *testing.T) {
	cfg := &config.Config{
		Repositories: []config.Repository{
			{Name: "test", URL: "https://github.com/test/repo"},
		},
	}

	c := New(cfg, "test-token")

	assert.NotNil(t, c)
	assert.Equal(t, cfg, c.config)
	assert.NotNil(t, c.client)
	assert.NotNil(t, c.fetcher)
}

func TestConductor_Run_NoRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Repositories: []config.Repository{},
	}

	c := New(cfg, "test-token")
	// Replace the fetcher with a mock for testing
	c.fetcher = repo.NewMockFetcher(ctrl)

	ctx := context.Background()
	err := c.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no repositories configured")
}

func TestConductor_Run_WithRepositories_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Repositories: []config.Repository{
			{Name: "test", URL: "https://github.com/test/repo"},
		},
	}

	expectedResults := map[string][]byte{
		"go.mod": []byte("module github.com/test/repo\nrequire github.com/test/dep v1.0.0\n"),
	}

	mockFetcher := repo.NewMockFetcher(ctrl)
	mockFetcher.EXPECT().
		FetchRepositoryFiles(gomock.Any(), "https://github.com/test/repo", "main", "go.mod").
		Return(expectedResults, nil)

	mockGraphBuilder := depgraph.NewMockGraphBuilder(ctrl)
	mockGraph := map[string]*depgraph.Service{
		"github.com/test/repo": {
			ModulePath:   "github.com/test/repo",
			RepoURL:      "https://github.com/test/repo",
			Dependencies: map[string]depgraph.Dependency{},
		},
	}
	mockGraphBuilder.EXPECT().BuildGraph(gomock.Any()).Return(mockGraph, nil)

	mockVersionDetector := repo.NewMockVersionDetector(ctrl)
	mockVersionDetector.EXPECT().DetectAndSetCurrentVersions(gomock.Any(), gomock.Any(), mockGraph).Return(nil)

	c := New(cfg, "test-token")
	c.fetcher = mockFetcher
	c.graphBuilder = mockGraphBuilder
	c.versionDetector = mockVersionDetector

	ctx := context.Background()
	err := c.Run(ctx)

	assert.NoError(t, err)
}

func TestConductor_Run_WithMultipleRepositories_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Repositories: []config.Repository{
			{Name: "repo1", URL: "https://github.com/test/repo1"},
			{Name: "repo2", URL: "https://github.com/test/repo2"},
		},
	}

	mockFetcher := repo.NewMockFetcher(ctrl)
	mockFetcher.EXPECT().
		FetchRepositoryFiles(gomock.Any(), "https://github.com/test/repo1", "main", "go.mod").
		Return(map[string][]byte{"go.mod": []byte("module github.com/test/repo1")}, nil)
	mockFetcher.EXPECT().
		FetchRepositoryFiles(gomock.Any(), "https://github.com/test/repo2", "main", "go.mod").
		Return(map[string][]byte{"go.mod": []byte("module github.com/test/repo2")}, nil)

	mockGraphBuilder := depgraph.NewMockGraphBuilder(ctrl)
	mockGraph := map[string]*depgraph.Service{
		"github.com/test/repo1": {
			ModulePath:   "github.com/test/repo1",
			RepoURL:      "https://github.com/test/repo1",
			Dependencies: map[string]depgraph.Dependency{},
		},
		"github.com/test/repo2": {
			ModulePath:   "github.com/test/repo2",
			RepoURL:      "https://github.com/test/repo2",
			Dependencies: map[string]depgraph.Dependency{},
		},
	}
	mockGraphBuilder.EXPECT().BuildGraph(gomock.Any()).Return(mockGraph, nil)

	mockVersionDetector := repo.NewMockVersionDetector(ctrl)
	mockVersionDetector.EXPECT().DetectAndSetCurrentVersions(gomock.Any(), gomock.Any(), mockGraph).Return(nil)

	c := New(cfg, "test-token")
	c.fetcher = mockFetcher
	c.graphBuilder = mockGraphBuilder
	c.versionDetector = mockVersionDetector

	ctx := context.Background()
	err := c.Run(ctx)

	assert.NoError(t, err)
}

func TestConductor_Run_WithRepositories_FetchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		Repositories: []config.Repository{
			{Name: "test", URL: "https://github.com/test/repo"},
		},
	}

	mockFetcher := repo.NewMockFetcher(ctrl)
	mockFetcher.EXPECT().
		FetchRepositoryFiles(gomock.Any(), "https://github.com/test/repo", "main", "go.mod").
		Return(nil, assert.AnError)

	c := New(cfg, "test-token")
	// Replace the fetcher with a mock for testing
	c.fetcher = mockFetcher

	ctx := context.Background()
	err := c.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error fetching go.mod")
}
