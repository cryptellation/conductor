//go:generate go run go.uber.org/mock/mockgen@v0.2.0 -source=client.go -destination=mock_client.go -package=github
package github

import (
	"context"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

// Client defines the interface for interacting with GitHub.
type Client interface {
	GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error)
	ListTags(ctx context.Context, owner, repo string) ([]*github.RepositoryTag, error)
}

// client implements Client using go-github.
type client struct {
	gh *github.Client
}

// New creates a new GitHub client with the given token.
func New(token string) Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	gh := github.NewClient(oauth2.NewClient(context.Background(), ts))
	return &client{gh: gh}
}

// GetFileContent retrieves the content of a file from a GitHub repository.
func (c *client) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error) {
	fileContent, _, _, err := c.gh.Repositories.GetContents(
		ctx, owner, repo, path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, err
	}
	if fileContent == nil {
		return nil, nil
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

// ListTags retrieves the tags of a GitHub repository.
func (c *client) ListTags(ctx context.Context, owner, repo string) ([]*github.RepositoryTag, error) {
	tags, _, err := c.gh.Repositories.ListTags(ctx, owner, repo, nil)
	return tags, err
}
