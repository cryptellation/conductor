# Automated Merge Requests

This document outlines the Automated Merge Requests feature for the DepSync tool. This feature focuses on automatically creating, tracking, and merging pull requests for dependency updates.

## Current Status

**Feature Status**: ✅ Complete

All tasks in this feature have been completed and are ready for use.

## Feature Overview

The Automated Merge Requests feature provides end-to-end automation for dependency updates. It includes Dagger-based workflows for repository operations, merge request creation, status tracking, and automatic merging with CI/CD integration.

## Completed Tasks

### 2.1. Dagger Workflow

#### 2.1.1. Git Clone Implementation
- **Status**: ✅ Done
- **Description**: Implement a way to git clone a project with Dagger
- **Implementation Details**:
  - Implemented as Go function in `/pkg/dagger`
  - Uses Dagger Go SDK for shallow git clone
  - Returns `*dagger.Directory` for chaining
  - Default branch "main" (configurable constant)
  - Uses `GITHUB_TOKEN` for authentication
  - Fail fast on errors
  - Logs progress with zap logger
  - No submodules required
  - Part of `UpdateDependencyWorkflow` function
  - Accepts mismatches from `depsync.go`

#### 2.1.2. Dependency Update Implementation
- **Status**: ✅ Done
- **Description**: Implement a way to update a go dependency in the cloned project using Dagger
- **Implementation Details**:
  - New method on existing Dagger interface in `/pkg/adapters/dagger`
  - Uses `go get` to update single dependency
  - Function signature: `UpdateGoDependency(ctx context.Context, dir *dagger.Directory, modulePath, targetVersion string) (*dagger.Directory, error)`
  - Uses `github.com/Masterminds/semver/v3` for version parsing
  - Uses `go get module@version` format
  - Fail fast on errors
  - Captures and logs `go get` output
  - Returns updated directory for chaining
  - Called from `fixModules` method in `depsync.go`
  - Unit tests in `depsync_test.go`
  - Integration tests in `dagger_test.go`

#### 2.1.3. Commit and Push Implementation
- **Status**: ✅ Done
- **Description**: Implement a way to commit and push the change to a new branch using Dagger
- **Implementation Details**:
  - New method on existing Dagger interface
  - Function signature: `CommitAndPush(ctx context.Context, dir *dagger.Directory, modulePath, targetVersion string) (string, error)`
  - Branch naming: `depsync/update-<dependency>-<version>`
  - Commit message: `"fix(dependencies): update <dependency> to <version>"`
  - Git author configuration in config file
  - Branch names sanitized (invalid chars replaced with `-`)
  - Fail on existing branch (no retry)
  - Uses same GitHub token as cloning
  - Push immediately after commit
  - Uses `alpine/git` image
  - Called from `fixModules` method after `UpdateGoDependency`
  - Stops at first failure
  - Logs branch name and success status
  - Unit and integration tests included

#### 2.1.4. Branch Existence Check
- **Status**: ✅ Done
- **Description**: Check if the future created branch already exists
- **Implementation Details**:
  - New method on existing Dagger interface
  - Function signature: `CheckBranchExists(ctx context.Context, dir *dagger.Directory, modulePath, targetVersion, repoURL string) (bool, error)`
  - Check happens after `CloneRepo` but before `UpdateGoDependency`
  - Uses same branch naming convention as `CommitAndPush`
  - Uses same GitHub token for authentication
  - Returns boolean (true = exists, false = doesn't exist)
  - Logs warning if branch exists, skips to next dependency
  - Called from `updateDependency` method right after clone
  - Uses `alpine/git` image
  - Uses `git ls-remote --heads origin <branch-name>`
  - Determines existence by checking command output
  - Fail fast on git operation errors
  - Logs include service, dependency, version, repository URL
  - In `updateDependency`: calls `CheckBranchExists` after `CloneRepo`, skips if exists, continues if doesn't exist
  - Unit and integration tests included

### 2.2. MR Creation Logic

#### 2.2.1. Merge Request Creation
- **Status**: ✅ Done
- **Description**: Implement logic to create merge requests in affected repositories
- **Implementation Details**:
  - New method on existing GitHub adapter in `/pkg/adapters/github`
  - Function signature: `CreateMergeRequest(ctx context.Context, repoURL, sourceBranch, modulePath, targetVersion string) error`
  - MR title format: "[{git name}] Update {dependency} to {version}"
  - Separate `GenerateMRTitle` function for reuse
  - Auto-generated MR description with update details
  - Source branch: `depsync/update-{dependency}-{version}`
  - Target branch: constant (default "main")
  - Fail fast on MR creation errors
  - Modify `updateDependency` to return branch name
  - Called from new `manageMergeRequest` method
  - Creates one MR per dependency update
  - No return values needed
  - No configuration needed

### 2.3. MR Status Tracking

#### 2.3.1. Status Tracking and CI/CD Checks
- **Status**: ✅ Done
- **Description**: Track the status of created MRs and wait for checks to pass. Never bypass checks.
- **Implementation Details**:
  - Modified `CreateMergeRequest` to return PR number
  - Added new methods to GitHub client interface for PR details and check status
  - Non-blocking check in `manageMergeRequest` for CI/CD status
  - Check status once immediately after MR creation
  - Logs current state (running/passed/failed)
  - Continues to next services/dependencies if checks running
  - Logs failure and continues with other MRs if checks fail
  - Logs only pass/fail status, not detailed check information
  - Uses GitHub Checks API
  - Considers PR successful when all CI/CD checks pass

### 2.4. MR Merging

#### 2.4.2. Automatic Merging and Branch Cleanup
- **Status**: ✅ Done
- **Description**: After checking MR status, if checks are successful, merge the MR and delete the branch
- **Implementation Details**:
  - Renamed `checkAndLogCIStatus` to `checkAndMergeMR`
  - Added merge functionality when CI/CD checks pass
  - Uses squash merge strategy
  - Deletes branch immediately after successful merge
  - Continues with other MRs and logs failure if merge fails
  - New `MergeMergeRequest` method in existing GitHub adapter
  - Called from `checkAndMergeMR` on "passed" case
  - Uses hardcoded settings (no configuration needed)
  - Logs success/failure and branch deletion success/failure
  - Updated unit tests in `depsync_test.go`
  - Uses same commit message format: `"fix(dependencies): update <dependency> to <version>"`
  - Passes mismatch information to merge function

## Configuration

The feature requires GitHub authentication and git author configuration:

```yaml
git:
  author:
    name: "DepSync Bot"
    email: "depsync@example.com"

github:
  token: "${GITHUB_TOKEN}"
```

## Usage

The automated merge requests feature is used through the main DepSync command:

```bash
depsync update
```

This will:
1. Detect dependency inconsistencies
2. Clone affected repositories
3. Update dependencies
4. Create branches and commit changes
5. Create merge requests
6. Track CI/CD status
7. Merge successful MRs
8. Clean up branches

## Workflow

1. **Detection**: Identifies dependency mismatches
2. **Cloning**: Shallow clones repositories using Dagger
3. **Branch Check**: Verifies branch doesn't already exist
4. **Update**: Uses `go get` to update specific dependency
5. **Commit**: Creates commit with standardized message
6. **Push**: Pushes to new branch
7. **MR Creation**: Creates merge request with descriptive title
8. **Status Tracking**: Monitors CI/CD checks
9. **Merging**: Automatically merges when checks pass
10. **Cleanup**: Deletes feature branch

## Architecture

- **Dagger Integration**: Containerized workflow execution
- **GitHub API**: REST API integration for MR management
- **CI/CD Integration**: Checks API for status monitoring
- **Error Handling**: Fail-fast approach with detailed logging
- **Testing**: Mock-based unit tests and real integration tests

## Dependencies

- `dagger.io/dagger` - Containerized workflow execution
- `github.com/google/go-github/v57/github` - GitHub API client
- `go.uber.org/zap` - Structured logging
- `github.com/Masterminds/semver/v3` - Version parsing 