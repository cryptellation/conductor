//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=client.go -destination=mock.gen.go -package=github
package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cryptellation/depsync/pkg/adapters"
	"github.com/cryptellation/depsync/pkg/logging"
	"github.com/google/go-github/v55/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// DepSyncPRTitlePrefix is the prefix used for DepSync pull request titles.
const DepSyncPRTitlePrefix = "chores(depsync):"

// GetFileContentParams contains parameters for GetFileContent.
type GetFileContentParams struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

// CreateMergeRequestParams contains parameters for CreateMergeRequest.
type CreateMergeRequestParams struct {
	RepoURL       string
	SourceBranch  string
	ModulePath    string
	TargetVersion string
}

// CheckPullRequestExistsParams contains parameters for CheckPullRequestExists.
type CheckPullRequestExistsParams struct {
	RepoURL      string
	SourceBranch string
}

// GetPullRequestChecksParams contains parameters for GetPullRequestChecks.
type GetPullRequestChecksParams struct {
	RepoURL  string
	PRNumber int
}

// MergeMergeRequestParams contains parameters for MergeMergeRequest.
type MergeMergeRequestParams struct {
	RepoURL       string
	PRNumber      int
	ModulePath    string
	TargetVersion string
}

// DeleteBranchParams contains parameters for DeleteBranch.
type DeleteBranchParams struct {
	RepoURL    string
	BranchName string
}

type DeletePullRequestParams struct {
	RepoURL  string
	PRNumber int
}

// CheckMergeConflictsParams contains parameters for CheckMergeConflicts.
type CheckMergeConflictsParams struct {
	RepoURL  string
	PRNumber int
}

// CheckStatus represents the status of CI/CD checks for a pull request.
type CheckStatus struct {
	Status string // "running", "passed", "failed"
}

// Client defines the interface for interacting with GitHub.
type Client interface {
	GetFileContent(ctx context.Context, params GetFileContentParams) ([]byte, error)
	ListTags(ctx context.Context, owner, repo string) ([]*github.RepositoryTag, error)
	CreateMergeRequest(ctx context.Context, params CreateMergeRequestParams) (int, error)
	CheckPullRequestExists(ctx context.Context, params CheckPullRequestExistsParams) (int, error)
	GetPullRequestChecks(ctx context.Context, params GetPullRequestChecksParams) (*CheckStatus, error)
	MergeMergeRequest(ctx context.Context, params MergeMergeRequestParams) error
	DeleteBranch(ctx context.Context, params DeleteBranchParams) error
	DeletePullRequest(ctx context.Context, params DeletePullRequestParams) error
	CheckMergeConflicts(ctx context.Context, params CheckMergeConflictsParams) (bool, error)
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
func (c *client) GetFileContent(ctx context.Context, params GetFileContentParams) ([]byte, error) {
	fileContent, _, _, err := c.gh.Repositories.GetContents(
		ctx, params.Owner, params.Repo, params.Path,
		&github.RepositoryContentGetOptions{Ref: params.Ref},
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

// CreateMergeRequest creates a merge request in the specified repository.
func (c *client) CreateMergeRequest(ctx context.Context, params CreateMergeRequestParams) (int, error) {
	// Extract owner and repo from the repository URL
	// Format: https://github.com/owner/repo
	parts := strings.Split(strings.TrimPrefix(params.RepoURL, "https://"), "/")
	if len(parts) != 3 {
		return -1, fmt.Errorf("invalid repository URL format: %s", params.RepoURL)
	}
	owner := parts[1]
	repo := parts[2]

	// Generate MR title and description
	title := generateMRTitle(params.ModulePath, params.TargetVersion)
	description := generateMRDescription(params.ModulePath, params.TargetVersion)

	// Create the pull request
	pr := &github.NewPullRequest{
		Title: &title,
		Body:  &description,
		Head:  &params.SourceBranch,
		Base:  github.String("main"), // Using constant as specified
	}

	createdPR, _, err := c.gh.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		return -1, err
	}

	return *createdPR.Number, nil
}

// CheckPullRequestExists checks if a pull request already exists for the given branch.
// Returns the PR number if it exists, or -1 if it doesn't exist.
func (c *client) CheckPullRequestExists(ctx context.Context, params CheckPullRequestExistsParams) (int, error) {
	// Extract owner and repo from the repository URL
	// Format: https://github.com/owner/repo
	parts := strings.Split(strings.TrimPrefix(params.RepoURL, "https://"), "/")
	if len(parts) != 3 {
		return -1, fmt.Errorf("invalid repository URL format: %s", params.RepoURL)
	}
	owner := parts[1]
	repo := parts[2]

	// List pull requests with the specific head branch
	opts := &github.PullRequestListOptions{
		Head:  fmt.Sprintf("%s:%s", owner, params.SourceBranch),
		State: "open",
	}

	pulls, _, err := c.gh.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return -1, err
	}

	// If any pull requests are returned, return the first one's number
	if len(pulls) > 0 {
		return *pulls[0].Number, nil
	}

	// No pull request found
	return -1, nil
}

// GetPullRequestChecks gets the status of CI/CD checks for a pull request.
func (c *client) GetPullRequestChecks(ctx context.Context, params GetPullRequestChecksParams) (*CheckStatus, error) {
	owner, repo, err := extractOwnerAndRepo(params.RepoURL)
	if err != nil {
		return nil, err
	}

	// Get the pull request to find the head SHA
	pr, _, err := c.gh.PullRequests.Get(ctx, owner, repo, params.PRNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	// Get check runs for the head commit
	checkRuns, _, err := c.gh.Checks.ListCheckRunsForRef(ctx, owner, repo, *pr.Head.SHA, &github.ListCheckRunsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get check runs: %w", err)
	}

	return determineCheckStatus(checkRuns.CheckRuns), nil
}

// checkMergeConflictsWithRetry performs the actual merge conflict check with retry logic.
func (c *client) checkMergeConflictsWithRetry(ctx context.Context, owner, repo string, prNumber int) (bool, error) {
	maxRetries := 5
	baseDelay := time.Second * 2

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get the pull request to check for conflicts
		pr, _, err := c.gh.PullRequests.Get(ctx, owner, repo, prNumber)
		if err != nil {
			return false, fmt.Errorf("failed to get pull request: %w", err)
		}

		// Check mergeable state
		mergeableState := pr.GetMergeableState()

		// GitHub API can return null or unknown when mergeability hasn't been computed yet
		// We need to handle these cases
		var hasConflicts bool

		switch mergeableState {
		case "clean":
			// No conflicts
			hasConflicts = false
		case "conflicting":
			// Has actual merge conflicts
			hasConflicts = true
		case "unstable", "dirty", "blocked":
			// These states don't indicate merge conflicts:
			// - "unstable": Usually failing CI checks
			// - "dirty": Usually needs rebase
			// - "blocked": Usually branch protection rules
			hasConflicts = false
		case "unknown":
			// GitHub hasn't computed mergeability yet, wait and retry
			if attempt < maxRetries-1 {
				delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
				logging.L().Info("Mergeability not yet computed for PR",
					zap.Int("pr_number", prNumber),
					zap.Int("attempt", attempt+1),
					zap.Int("max_retries", maxRetries),
					zap.Duration("delay", delay),
				)
				time.Sleep(delay)
				continue
			}
			// If we've exhausted all retries, return an error
			return false, fmt.Errorf("mergeability not yet computed after %d retries, please try again later", maxRetries)
		default:
			// Unknown state, return an error
			return false, fmt.Errorf("unknown mergeable state: %q", mergeableState)
		}

		return hasConflicts, nil
	}

	// This should never be reached, but just in case
	return false, fmt.Errorf("unexpected retry loop exit")
}

// CheckMergeConflicts checks if a pull request has merge conflicts.
func (c *client) CheckMergeConflicts(ctx context.Context,
	params CheckMergeConflictsParams) (bool, error) {
	owner, repo, err := extractOwnerAndRepo(params.RepoURL)
	if err != nil {
		return false, err
	}

	return c.checkMergeConflictsWithRetry(ctx, owner, repo, params.PRNumber)
}

// MergeMergeRequest merges a pull request.
func (c *client) MergeMergeRequest(ctx context.Context, params MergeMergeRequestParams) error {
	owner, repo, err := extractOwnerAndRepo(params.RepoURL)
	if err != nil {
		return err
	}

	// Merge the pull request with squash strategy
	// The commit message will be the PR title (which is already in the correct format)
	_, _, err = c.gh.PullRequests.Merge(ctx, owner, repo, params.PRNumber, "", &github.PullRequestOptions{
		MergeMethod: "squash",
	})
	if err != nil {
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	return nil
}

// DeleteBranch deletes a branch from a GitHub repository.
func (c *client) DeleteBranch(ctx context.Context, params DeleteBranchParams) error {
	owner, repo, err := extractOwnerAndRepo(params.RepoURL)
	if err != nil {
		return err
	}

	// Delete the branch using GitHub API
	_, err = c.gh.Git.DeleteRef(ctx, owner, repo, fmt.Sprintf("refs/heads/%s", params.BranchName))
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", params.BranchName, err)
	}

	return nil
}

// DeletePullRequest closes a pull request in a GitHub repository.
func (c *client) DeletePullRequest(ctx context.Context, params DeletePullRequestParams) error {
	owner, repo, err := extractOwnerAndRepo(params.RepoURL)
	if err != nil {
		return err
	}

	// Close the pull request by updating its state to "closed"
	_, _, err = c.gh.PullRequests.Edit(ctx, owner, repo, params.PRNumber, &github.PullRequest{
		State: github.String("closed"),
	})
	if err != nil {
		return fmt.Errorf("failed to close pull request %d: %w", params.PRNumber, err)
	}

	return nil
}

// extractOwnerAndRepo extracts owner and repo from a GitHub URL.
func extractOwnerAndRepo(repoURL string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(repoURL, "https://"), "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid repository URL format: %s", repoURL)
	}
	return parts[1], parts[2], nil
}

// determineCheckStatus determines the overall status of check runs.
func determineCheckStatus(checkRuns []*github.CheckRun) *CheckStatus {
	if len(checkRuns) == 0 {
		// No checks found, consider as running
		return &CheckStatus{Status: "running"}
	}

	// Check if any checks are still running
	for _, check := range checkRuns {
		if *check.Status == "in_progress" || *check.Status == "queued" {
			return &CheckStatus{Status: "running"}
		}
	}

	// Check if any checks failed
	for _, check := range checkRuns {
		if *check.Conclusion == "failure" || *check.Conclusion == "cancelled" || *check.Conclusion == "timed_out" {
			return &CheckStatus{Status: "failed"}
		}
	}

	// All checks passed
	return &CheckStatus{Status: "passed"}
}

// generateMRTitle generates the title for a merge request.
func generateMRTitle(modulePath, targetVersion string) string {
	return adapters.FormatCommitMessage(modulePath, targetVersion)
}

// generateMRDescription generates the description for a merge request.
func generateMRDescription(modulePath, targetVersion string) string {
	return fmt.Sprintf(`## Dependency Update

This merge request updates the dependency **%s** to version **%s**.

### Changes
- Updated dependency: `+"`%s`"+`
- New version: `+"`%s`"+`

This update was automatically generated by DepSync.`, modulePath, targetVersion, modulePath, targetVersion)
}
