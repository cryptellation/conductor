package github

import (
	"context"
	"os"
	"testing"
)

func TestGetFileContent(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set; skipping integration test.")
	}

	client := New(token)
	ctx := context.Background()

	owner := "octocat"
	repo := "Hello-World"
	path := "README"
	ref := "master"

	content, err := client.GetFileContent(ctx, owner, repo, path, ref)
	if err != nil {
		t.Fatalf("failed to get file content: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("expected file content, got empty result")
	}
}
