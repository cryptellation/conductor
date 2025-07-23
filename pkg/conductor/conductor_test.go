package conductor

import (
	"context"
	"testing"

	"github.com/lerenn/conductor/pkg/config"
	"github.com/lerenn/conductor/pkg/repofetcher"
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
	c.fetcher = repofetcher.NewMockFetcher(ctrl)

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
		"README.md": []byte("# Test Repository"),
		"LICENSE":   []byte("MIT License"),
	}

	mockFetcher := repofetcher.NewMockFetcher(ctrl)
	mockFetcher.EXPECT().
		FetchRepositoryFiles(gomock.Any(), "https://github.com/test/repo", "main", "README.md", "LICENSE").
		Return(expectedResults, nil)

	c := New(cfg, "test-token")
	// Replace the fetcher with a mock for testing
	c.fetcher = mockFetcher

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

	mockFetcher := repofetcher.NewMockFetcher(ctrl)
	mockFetcher.EXPECT().
		FetchRepositoryFiles(gomock.Any(), "https://github.com/test/repo", "main", "README.md", "LICENSE").
		Return(nil, assert.AnError)

	c := New(cfg, "test-token")
	// Replace the fetcher with a mock for testing
	c.fetcher = mockFetcher

	ctx := context.Background()
	err := c.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error fetching repository files")
}
