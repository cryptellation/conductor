package repo

import (
	"context"
	"errors"

	"github.com/lerenn/conductor/pkg/adapters/github"
)

//go:generate go run go.uber.org/mock/mockgen@v0.2.0 -source=fetcher.go -destination=mock_fetcher.gen.go -package=repo

// ErrInvalidRepoURL is returned when the repository URL cannot be parsed.
var ErrInvalidRepoURL = errors.New("invalid repository URL")

// FilesFetcher defines the interface for fetching repository files.
type FilesFetcher interface {
	Fetch(ctx context.Context, repoURL, ref string, files ...string) (map[string][]byte, error)
}

// fetcher fetches content from configured repositories using the GitHub adapter.
type fetcher struct {
	client github.Client
}

// Ensure fetcher implements Fetcher.
var _ FilesFetcher = (*fetcher)(nil)

func NewFilesFetcher(client github.Client) FilesFetcher {
	return &fetcher{client: client}
}

// Fetch fetches the content of the given files from the specified repository URL and ref.
func (f *fetcher) Fetch(
	ctx context.Context,
	repoURL, ref string,
	files ...string,
) (map[string][]byte, error) {
	owner, name := parseOwnerAndRepo(repoURL)
	if owner == "" || name == "" {
		return nil, ErrInvalidRepoURL
	}
	results := make(map[string][]byte)
	for _, file := range files {
		content, err := f.client.GetFileContent(ctx, owner, name, file, ref)
		if err != nil {
			return nil, err
		}
		results[file] = content
	}
	return results, nil
}
