package repofetcher

import (
	"context"

	"github.com/lerenn/conductor/pkg/adapters/github"
	"github.com/lerenn/conductor/pkg/config"
)

// RepositoryContentFetcher fetches content from configured repositories using the GitHub adapter.
type RepositoryContentFetcher struct {
	client github.Client
}

func NewRepositoryContentFetcher(client github.Client) *RepositoryContentFetcher {
	return &RepositoryContentFetcher{client: client}
}

// FetchAllRepositoriesContent fetches the content of a given file (e.g., README.md) from all configured repositories.
// The 'path' and 'ref' parameters specify which file and ref to fetch (e.g., "README.md", "main").
func (f *RepositoryContentFetcher) FetchAllRepositoriesContent(
	ctx context.Context,
	cfg *config.Config,
	path, ref string,
) (map[string][]byte, error) {
	results := make(map[string][]byte)
	for _, repo := range cfg.Repositories {
		owner, name := parseOwnerAndRepo(repo.URL)
		if owner == "" || name == "" {
			continue // skip invalid URLs
		}
		content, err := f.client.GetFileContent(ctx, owner, name, path, ref)
		if err != nil {
			results[repo.Name] = nil // or handle error differently
			continue
		}
		results[repo.Name] = content
	}
	return results, nil
}

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
