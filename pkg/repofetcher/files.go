package repofetcher

import (
	"context"
	"errors"

	"github.com/lerenn/conductor/pkg/adapters/github"
)

// RepositoryFilesFetcher fetches content from configured repositories using the GitHub adapter.
type RepositoryFilesFetcher struct {
	client github.Client
}

func NewRepositoryFilesFetcher(client github.Client) *RepositoryFilesFetcher {
	return &RepositoryFilesFetcher{client: client}
}

// FetchRepositoryFiles fetches the content of the given files from the specified repository URL and ref.
func (f *RepositoryFilesFetcher) FetchRepositoryFiles(
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

// ErrInvalidRepoURL is returned when the repository URL cannot be parsed.
var ErrInvalidRepoURL = errors.New("invalid repository URL")

// parseOwnerAndRepo extracts the owner and repo name from a GitHub URL.
func parseOwnerAndRepo(url string) (owner, repo string) {
	// Example: https://github.com/example/testrepo1.git
	// Should return ("example", "testrepo1")
	// This is a simple implementation; can be improved for edge cases.
	const prefix = "github.com/"
	idx := -1
	for i := 0; i < len(url)-len(prefix); i++ {
		if url[i:i+len(prefix)] == prefix {
			idx = i + len(prefix)
			break
		}
	}
	if idx == -1 {
		return "", ""
	}
	rest := url[idx:]
	if len(rest) == 0 {
		return "", ""
	}
	// Remove .git suffix if present
	if rest[len(rest)-4:] == ".git" {
		rest = rest[:len(rest)-4]
	}
	parts := make([]string, 0, 2)
	for _, p := range rest {
		if p == '/' {
			parts = append(parts, "")
			continue
		}
		if len(parts) == 0 {
			parts = append(parts, "")
		}
		parts[len(parts)-1] += string(p)
	}
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
