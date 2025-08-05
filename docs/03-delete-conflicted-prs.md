# Delete Conflicted PRs

## Overview

The delete conflicted PRs feature automatically deletes pull requests and their associated branches when conflicts are detected in DepSync PRs. This feature is designed to handle scenarios where multiple services update their dependencies simultaneously, and conflicts arise between different dependency update PRs.

### Problem Statement

When multiple dependency updates are happening simultaneously across different services, conflicts can arise when:

1. Service A updates dependency X to version 1.2.0
2. Service B updates dependency Y to version 2.1.0  
3. Both services depend on each other
4. Service A's PR gets merged first
5. Service B's PR now has conflicts because the merged changes from Service A affected the same lines in `go.mod`

Without automatic deletion, this would require manual intervention to resolve the conflicts, delaying the dependency update process.

### Solution

The delete conflicted PRs feature automatically detects when a DepSync PR has any conflicts, then deletes the PR and its branch to allow for a fresh start on the next depsync cycle:

1. Detecting conflicts in existing DepSync PRs using GitHub API
2. Deleting the conflicted PR and its associated branch
3. Waiting for the next depsync cycle to recreate the PR with updated dependencies

## Requirements

### Functional Requirements

- **Conflict Detection**: Use GitHub API to detect merge conflicts in DepSync PRs using the `mergeableState` field
- **Scope Limitation**: Delete any DepSync PR that has conflicts, regardless of which files are affected
- **Depsync Only**: Only apply deletion to PRs created by depsync (identified by `[DepSync]` prefix in PR title using shared constant)
- **PR and Branch Deletion**: Delete both the pull request and its associated branch
- **GitHub Integration**: Execute PR and branch deletion through GitHub API using existing functions where possible
- **Configuration**: Respect existing configuration options and add new `delete_conflicted_prs` option as global setting only (default: true)
- **Error Handling**: Stop execution and return error if deletion operations fail
- **Logging**: Provide detailed logging for debugging and monitoring using existing logging patterns

### Non-Functional Requirements

- **Performance**: Deletion should complete within the same timeout limits as other operations (use default values)
- **Reliability**: Should handle edge cases gracefully without breaking the workflow
- **Security**: Use existing authentication and authorization mechanisms
- **Monitoring**: Provide clear logging for audit trails and debugging
- **Serialization**: Deletion operations should be serialized (not concurrent)

## Implementation

### Architecture

The delete conflicted PRs feature integrates with the existing depsync workflow as a separate step that runs when PRs already exist (not when they are created).

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   PR Exists     │───▶│  Conflict Check  │───▶│  Delete PR      │
│                 │    │  (GitHub API)    │    │  (GitHub API)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Configuration

Add new configuration option to the depsync configuration as global setting only:

```yaml
# depsync.yaml
delete_conflicted_prs: true  # Global setting - enable/disable deletion of conflicted PRs (default: true)

repositories:
  - "service-a"
  - "service-b"

git:
  author:
    name: "DepSync Bot"
    email: "depsync@example.com"
```

### Workflow Integration

1. **PR Exists**: Check existing PRs for conflicts in `manageMergeRequest` function
2. **Conflict Detection**: Use GitHub API to check if PR has any conflicts using the `mergeableState` field
3. **Deletion Execution**: If conflicts detected and feature enabled, execute deletion as separate function to avoid making `manageMergeRequest` too long
4. **Error Handling**: If deletion fails, stop immediately and return error from the main depsync function (log error before returning)
5. **Continue Workflow**: Resume normal PR processing after deletion (but don't create new PR - wait for next depsync run)

### Integration Pattern in manageMergeRequest

The deletion logic integrates into the existing `manageMergeRequest` function using the following pattern:

```go
if prNumber == -1 {
    // existing PR creation logic
} else {
    // Conflict check and deletion
    if deleted {
        return nil  // Skip checkAndMergeMR call
    }
}

// existing checkAndMergeMR call
```

This ensures that:
- If no PR exists, create it as normal
- If PR exists, check for conflicts and perform deletion if needed
- If deletion was performed, skip the normal checkAndMergeMR call
- If no deletion was needed, continue with normal checkAndMergeMR
- If deletion fails, return the error immediately

### GitHub API Integration

Use the GitHub adapter (`pkg/adapters/github`) to:

1. **Check PR Status**: Use GitHub API to check if PR has conflicts using the `mergeableState` field from the PR object
2. **Verify PR Source**: Ensure PR is from depsync by checking `[DepSync]` prefix in PR title using shared constant
3. **Delete PR and Branch**: Use GitHub API to delete both the pull request and its associated branch

### GitHub API Operations

Execute the following operations through GitHub API:

1. **Check Conflicts**: Use existing `CheckMergeConflicts` function to detect conflicts using `mergeableState`
2. **Delete PR**: Use new `DeletePullRequest` function to close the pull request
3. **Delete Branch**: Use existing `DeleteBranch` function to delete the associated branch

### Error Handling

The system should handle the following error scenarios:

1. **GitHub API Errors**: Log error and continue without deletion
2. **Deletion Failures**: Stop immediately and return error from the main depsync function (log error before returning)
3. **Configuration Errors**: Log warning and disable deletion

### Logging

Provide detailed logging for:

- Conflict detection results
- Deletion execution steps
- Error conditions and resolutions
- Performance metrics

Use existing logging patterns (`logging.C(ctx)`) with same log level conventions as other operations.

## Examples

### Configuration Example

```yaml
# depsync.yaml
delete_conflicted_prs: true  # Global setting (default: true)

repositories:
  - "service-a"
  - "service-b"

git:
  author:
    name: "DepSync Bot"
    email: "depsync@example.com"
```

### Workflow Example

1. **Initial State**: Two PRs updating dependencies
   - PR #123: Service A updates dependency X to v1.2.0
   - PR #124: Service B updates dependency Y to v2.1.0

2. **First Merge**: PR #123 gets merged successfully

3. **Conflict Detection**: PR #124 now has conflicts in `go.mod`:
   ```
   <<<<<<< HEAD
   require github.com/example/dep-x v1.2.0
   require github.com/example/dep-y v1.0.0
   =======
   require github.com/example/dep-x v1.1.0
   require github.com/example/dep-y v2.1.0
   >>>>>>> feature/update-dependency-y
   ```

4. **Deletion Execution**:
   - System detects conflict using GitHub API `mergeableState` field
   - Uses existing `CheckMergeConflicts` function to confirm conflicts
   - Uses new `DeletePullRequest` function to close the PR
   - Uses existing `DeleteBranch` function to delete the branch
   - Logs deletion completion

5. **Result**: PR #124 is deleted, and a new PR will be created on the next depsync cycle

### Error Example

If deletion fails due to GitHub API issues:

```
ERROR: Failed to delete conflicted PR, stopping depsync process
Output: 
  DELETE /repos/owner/repo/pulls/124: 404 Not Found

Action: Stopping immediately and returning error from main depsync function (logged before returning)
```

## Integration with Existing Features

This feature integrates with existing depsync features:

- **Dependency Management**: Works with the dependency update workflow described in [01-dependency-management.md](01-dependency-management.md)
- **Automated Merge Requests**: Integrates with the automated PR system described in [02-automated-merge-requests.md](02-automated-merge-requests.md)
- **Integration Management**: Follows the same configuration and monitoring patterns as [03-integration-management.md](03-integration-management.md)

## Future Enhancements

Potential future improvements:

1. **Conflict Prediction**: Predict potential conflicts before they occur
2. **Batch Processing**: Handle multiple conflicting PRs simultaneously
3. **Conflict Prevention**: Coordinate dependency updates to minimize conflicts
4. **Advanced Resolution**: Support for more complex conflict resolution strategies

## Conclusion

The delete conflicted PRs feature provides an automated solution for handling dependency conflicts by removing conflicted PRs and allowing fresh PRs to be created on the next cycle, reducing manual intervention and accelerating the dependency update process while maintaining the reliability and safety of the depsync workflow. 