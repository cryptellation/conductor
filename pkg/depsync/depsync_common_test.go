//go:build unit
// +build unit

package depsync

import (
	"testing"

	"github.com/cryptellation/depsync/pkg/adapters/dagger"
	"github.com/cryptellation/depsync/pkg/adapters/github"
	"github.com/cryptellation/depsync/pkg/config"
	"github.com/cryptellation/depsync/pkg/depgraph"
	"github.com/cryptellation/depsync/pkg/repo"
	"go.uber.org/mock/gomock"
)

// TestDepSync contains all the mocks and the depsync instance for testing
type TestDepSync struct {
	DepSync             *DepSync
	MockController      *gomock.Controller
	MockFetcher         *repo.MockFilesFetcher
	MockGraphBuilder    *depgraph.MockGraphBuilder
	MockVersionDetector *repo.MockVersionDetector
	MockChecker         *depgraph.MockInconsistencyChecker
	MockDagger          *dagger.MockDagger
	MockGitHubClient    *github.MockClient
}

// newTestDepSync creates a TestDepSync instance with all mocked dependencies
func newTestDepSync(t *testing.T, cfg *config.Config) *TestDepSync {
	ctrl := gomock.NewController(t)

	// Create all mocks
	mockFetcher := repo.NewMockFilesFetcher(ctrl)
	mockGraphBuilder := depgraph.NewMockGraphBuilder(ctrl)
	mockVersionDetector := repo.NewMockVersionDetector(ctrl)
	mockChecker := depgraph.NewMockInconsistencyChecker(ctrl)
	mockDagger := dagger.NewMockDagger(ctrl)
	mockGitHubClient := github.NewMockClient(ctrl)

	// Set up default expectations
	mockDagger.EXPECT().Close().Return(nil)

	// Create DepSync directly, avoiding New() which requires Docker
	c := &DepSync{
		config:          cfg,
		client:          mockGitHubClient,
		fetcher:         mockFetcher,
		graphBuilder:    mockGraphBuilder,
		versionDetector: mockVersionDetector,
		checker:         mockChecker,
		dagger:          mockDagger,
	}

	return &TestDepSync{
		DepSync:             c,
		MockController:      ctrl,
		MockFetcher:         mockFetcher,
		MockGraphBuilder:    mockGraphBuilder,
		MockVersionDetector: mockVersionDetector,
		MockChecker:         mockChecker,
		MockDagger:          mockDagger,
		MockGitHubClient:    mockGitHubClient,
	}
} 