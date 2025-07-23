package repofetcher

import (
	"context"
	"testing"

	"conductor/pkg/adapters/github"
	"conductor/pkg/config"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate go run go.uber.org/mock/mockgen@v0.2.0 -source=../adapters/github/client.go -destination=../adapters/github/mock_client.go -package=github

func TestFetchAllRepositoriesContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockClient := github.NewMockClient(ctrl)
	fetcher := NewRepositoryContentFetcher(mockClient)

	cfg := &config.Config{
		Repositories: []config.Repository{
			{Name: "repo1", URL: "https://github.com/owner1/repo1.git"},
			{Name: "repo2", URL: "https://github.com/owner2/repo2.git"},
		},
	}

	ctx := context.Background()
	mockClient.EXPECT().GetFileContent(ctx, "owner1", "repo1", "README.md", "main").Return([]byte("content1"), nil)
	mockClient.EXPECT().GetFileContent(ctx, "owner2", "repo2", "README.md", "main").Return([]byte("content2"), nil)

	results, err := fetcher.FetchAllRepositoriesContent(ctx, cfg, "README.md", "main")
	require.NoError(t, err)
	require.Equal(t, map[string][]byte{
		"repo1": []byte("content1"),
		"repo2": []byte("content2"),
	}, results)
}
