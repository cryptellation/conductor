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

func TestListTags(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set; skipping integration test.")
	}

	client := New(token)
	ctx := context.Background()

	owner := "kubernetes"
	repo := "kubernetes"

	tags, err := client.ListTags(ctx, owner, repo)
	if err != nil {
		t.Fatalf("failed to list tags: %v", err)
	}
	if len(tags) == 0 {
		t.Errorf("expected at least one tag, got none")
	}
}
