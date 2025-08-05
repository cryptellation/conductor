# Dependency Management

This document outlines the Dependency Management feature for the DepSync tool. This feature focuses on detecting and managing dependency inconsistencies across repositories.

## Current Status

**Feature Status**: ✅ Complete

All tasks in this feature have been completed and are ready for use.

## Feature Overview

The Dependency Management feature provides comprehensive dependency tracking and inconsistency detection across multiple repositories. It includes configuration management, repository discovery, dependency graph construction, version detection, and inconsistency identification.

## Completed Tasks

### 1.1. Define Configuration Format
- **Status**: ✅ Done
- **Description**: Specify how repositories, services, and libraries are listed in a YAML configuration file.

### 1.2. Configuration Loading
- **Status**: ✅ Done
- **Description**: Use Viper to create a `main.go` under `cmd/depsync` that loads the YAML config.

### 1.3. GitHub Client Adapter
- **Status**: ✅ Done
- **Description**: Create an adapter under `internal/adapter` that proxies the GitHub Go client, allowing for interface abstraction and testing.
- **Subtasks**:
  - 1.3.1: Create the adapter ✅
  - 1.3.2: Add test to retrieve a file from a repository ✅
  - 1.3.3: Add test to retrieve repository tags ✅
  - 1.3.4: Add tests to Dagger module ✅

### 1.4. Repository Discovery
- **Status**: ✅ Done
- **Description**: Use the configuration and the adapter to get the content of the configured repositories. Business logic written in `internal/core`, uses Github adapter via interface with dependency injection, used by command in `cmd/depsync`. Uses Uber gomock to mock the Github adapter.

### 1.5. Dependency Graph Construction
- **Status**: ✅ Done
- **Subtasks**:
  - 1.5.1: Pull go.mod files from repositories listed in configuration ✅
  - 1.5.2: Build dependency graph using versioning ✅
    - Receives map of repository module paths to go.mod contents
    - Outputs map of module path to Service struct
    - Service contains: module path, repository URL, dependencies map
    - No duplicate Service structs for same module path
    - Builder in own package under /pkg
    - Tests provided for builder
    - /pkg/depsync updated to use builder

### 1.6. Version Detection
- **Status**: ✅ Done
- **Description**: Implement version detection using the GitHub adapter to get tags and detect the last version of each dependency.
- **Implementation Details**:
  - Only semantic versioning tags considered (e.g., v1.2.3)
  - Non-semver tags ignored
  - Pre-release tags ignored
  - Only repositories/services listed in config processed
  - Service struct updated with CurrentVersion field
  - Simple print output
  - Repositories with no tags ignored
  - Fail fast on errors
  - Tests use mocks only
  - Logic in pkg/repofetcher

### 1.7. Inconsistency Detection
- **Status**: ✅ Done
- **Description**: Detect inconsistencies between the last version of a dependency and the version used in libraries/services on all repositories.
- **Implementation Details**:
  - Only direct dependencies checked
  - Inconsistency: latest version > version used by service
  - Only semantic versioning tags considered
  - Uses common Go semantic versioning library
  - Implemented as InconsistencyChecker in pkg/depgraph
  - Check function returns map[string]map[string]Mismatch
  - Mismatch struct has Actual and Latest version fields
  - Output only contains mismatched dependencies
  - Dependencies with no latest version skipped
  - Fail fast on errors
  - Tests use mocks only
  - Used by pkg/depsync/depsync.go

## Configuration

The feature uses a YAML configuration file to define repositories and services:

```yaml
repositories:
  - name: "service-a"
    url: "https://github.com/org/service-a"
  - name: "service-b" 
    url: "https://github.com/org/service-b"

git:
  author:
    name: "DepSync Bot"
    email: "depsync@example.com"
```

## Usage

The dependency management feature is used through the main DepSync command:

```bash
depsync check
```

This will:
1. Load configuration
2. Discover repositories
3. Build dependency graph
4. Detect versions
5. Identify inconsistencies
6. Display results

## Architecture

- **Configuration**: Viper-based YAML loading
- **GitHub Integration**: Adapter pattern with interface abstraction
- **Dependency Graph**: Custom builder in pkg/depgraph
- **Version Detection**: GitHub tags API integration
- **Inconsistency Detection**: Semantic version comparison
- **Testing**: Mock-based unit tests with Uber gomock

## Dependencies

- `github.com/spf13/viper` - Configuration management
- `github.com/golang/mock` - Mock generation
- `github.com/Masterminds/semver/v3` - Version comparison
- `go.uber.org/zap` - Logging 