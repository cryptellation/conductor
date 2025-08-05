package adapters

import "fmt"

// DepSyncCommitPrefix is the prefix used for DepSync commit messages.
const DepSyncCommitPrefix = "chores(depsync):"

// FormatCommitMessage formats a commit message for dependency updates.
func FormatCommitMessage(modulePath, targetVersion string) string {
	return fmt.Sprintf("%s update %s to %s", DepSyncCommitPrefix, modulePath, targetVersion)
}
