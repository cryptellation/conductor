package repofetcher

import (
	"context"
	"errors"
	"testing"

	"github.com/lerenn/conductor/pkg/adapters/github"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFetchRepositoryFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockClient := github.NewMockClient(ctrl)
	fetcher := New(mockClient)

	repoURL := "https://github.com/owner1/repo1.git"
	ctx := context.Background()
	mockClient.EXPECT().GetFileContent(ctx, "owner1", "repo1", "README.md", "main").Return([]byte("content1"), nil)
	mockClient.EXPECT().GetFileContent(ctx, "owner1", "repo1", "LICENSE", "main").Return([]byte("content2"), nil)

	results, err := fetcher.FetchRepositoryFiles(ctx, repoURL, "main", "README.md", "LICENSE")
	require.NoError(t, err)
	require.Equal(t, map[string][]byte{
		"README.md": []byte("content1"),
		"LICENSE":   []byte("content2"),
	}, results)
}

func TestFetchRepositoryFiles_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockClient := github.NewMockClient(ctrl)
	fetcher := New(mockClient)

	ctx := context.Background()
	_, err := fetcher.FetchRepositoryFiles(ctx, "invalid-url", "main", "README.md")
	require.ErrorIs(t, err, ErrInvalidRepoURL)
}

func TestFetchRepositoryFiles_ErrorOnFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockClient := github.NewMockClient(ctrl)
	fetcher := New(mockClient)

	repoURL := "https://github.com/owner1/repo1.git"
	ctx := context.Background()
	mockClient.EXPECT().GetFileContent(ctx, "owner1", "repo1", "README.md", "main").Return(nil, errors.New("fetch error"))

	_, err := fetcher.FetchRepositoryFiles(ctx, repoURL, "main", "README.md")
	require.Error(t, err)
}
