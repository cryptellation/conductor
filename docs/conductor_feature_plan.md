# Conductor Feature Plan

This document outlines the actionable development plan for the Conductor tool. Only tasks that are 'to do' or 'in progress' are listed in the main sections. Completed tasks are moved to the end for reference. Use the 'Current Focus' section below to quickly identify the next steps.

---

## Current Focus
- 1.7.1 Detect inconsistencies between the last version of a dependency and the version used in libraries/services on all repositories, based on the dependency graph and version detection. (Status: to do)
- 2.1.1 Integrate with GitHub (and/or other platforms) for API access. (Status: )
- 2.1.2 Handle authentication and permissions. (Status: )
- 2.2.1 Implement logic to create merge requests in affected repositories. (Status: )
- 2.2.2 Populate MRs with relevant changes (e.g., dependency version bumps). (Status: )
- 2.3.1 Track the status of created MRs (open, merged, closed, etc.). (Status: )
- 2.3.2 Notify or log progress for each MR. (Status: )
- 2.4.1 Handle failed MR creations or API errors. (Status: )
- 2.4.2 Implement retry and notification mechanisms. (Status: )
- 3.1.1 Identify and support integration repositories (e.g., Docker Compose, Helm Charts). (Status: )
- 3.1.2 Parse and update integration configuration files. (Status: )
- 3.2.1 Create a merge request in the integration repository once all dependent MRs are merged. (Status: )
- 3.2.2 Ensure the integration layer references the latest, compatible versions. (Status: )
- 3.3.1 Validate that all referenced versions are compatible. (Status: )
- 3.3.2 Optionally run integration tests or checks. (Status: )
- 3.4.1 Document the update process and results. (Status: )
- 3.4.2 Provide summary reports or dashboards for visibility. (Status: )

---

## 1. Dependency Management

- **1.6. Version Detection**
  - 1.6.1 Implement version detection using the GitHub adapter to get tags and detect the last version of each dependency.
    - Status: done
    - Implementation clarifications (2024-06-10):
      - Only semantic versioning tags are considered (e.g., v1.2.3). Non-semver tags are ignored.
      - Pre-release tags (e.g., v1.2.3-beta) are ignored for now.
      - Only repositories/services listed in the config (roots of the dependency graph) are processed.
      - The dependency graph (from depgraph) is the source of the map to update.
      - The Service struct should be updated to include a field for the latest version tag (field name: CurrentVersion).
      - The result is only displayed using a simple print statement (no special formatting).
      - Repositories with no tags are ignored (not an error).
      - Any error fetching tags (other than no tags) should cause the process to fail immediately (fail fast).
      - Only the happy path is tested (no edge cases for now).
      - Tests use only mocks (no real GitHub calls yet) and are placed in a new test file.
      - The logic should be implemented in new files under pkg/repofetcher.
      - Not exposed via CLI or as a public API yet.

- **1.7. Inconsistency Detection**
  - 1.7.1 Detect inconsistencies between the last version of a dependency and the version used in libraries/services on all repositories, based on the dependency graph and version detection.
    - Status: done
    - Implementation clarifications (2024-06-11):
      - Only repositories/services listed in the config (roots of the dependency graph) are checked for inconsistencies.
      - Only direct dependencies are checked (not transitive dependencies).
      - An inconsistency is defined as: the latest version of a dependency (as detected) is greater than the version used by a service/library in its direct dependencies (any difference, not just major/minor).
      - Only semantic versioning tags are considered (e.g., v1.2.3). Non-semver and pre-release tags are ignored.
      - Use the most common Go semantic versioning library for version comparison.
      - The logic should be implemented in a new file under `pkg/depgraph` as a new struct with its own interface and mock, named `InconsistencyChecker`, with a function `Check`.
      - The `Check` function should take the dependency graph (`map[string]*Service`) as parameter and return a `map[string]map[string]Mismatch`, where the first string is the service Go module path, the second is its dependency Go module path, and `Mismatch` is a struct with `Actual` and `Latest` version fields (no additional fields).
      - The output should only contain mismatched dependencies (no output for up-to-date dependencies).
      - If a dependency has no detected latest version (e.g., no tags), it is skipped (not an error).
      - Any error during the process should cause a fail-fast (immediate failure) with a descriptive error message.
      - Tests should use only mocks (generated with Uber gomock) and cover the happy path (no edge cases required).
      - The logic should be internal and used by `pkg/conductor/conductor.go` to print mismatches (simple print output only).

---

## 2. Automated Merge Requests

- **2.1. Dagger workflow**
  - 2.1.1 Implement a way to git clone a project with Dagger
    - Status: in progress
    - Implementation clarifications (2024-06-12):
      - Implement as a Go function in `/pkg/dagger` (not as a Dagger module).
      - The function should use the Dagger Go SDK to perform a shallow git clone of a given repository URL.
      - The function signature should return a Dagger Directory (`*dagger.Directory`) for chaining with other Dagger steps.
      - The branch to clone should default to "main" but be defined as a constant in `/pkg/dagger` for easy future changes.
      - Use the `GITHUB_TOKEN` provided to Conductor for authentication (supporting private repos). The token should be passed explicitly to the workflow and set as a Dagger secret if possible.
      - The function should fail fast on any error.
      - The function should log progress using the same logger as the rest of Conductor (e.g., zap). Only Conductor logs should be shown; Dagger-internal logs should be suppressed or not shown to the user.
      - No submodules required for now.
      - The function should be designed for integration into the Conductor workflow, not as a standalone CLI or Dagger module.
      - The cloned directory should be accessible for subsequent Dagger steps (e.g., updating dependencies, committing, pushing).
      - No special Git version or features required.
      - No need to export the directory to the host filesystem unless required by later steps.
      - Do not modify or touch any existing Dagger modules.
      - The `CloneRepo` function should be part of an `UpdateDependencyWorkflow` function in `/pkg/dagger`, which will also include steps for 2.2 and 2.3 in the future.
      - The `UpdateDependencyWorkflow` should accept the mismatches from `conductor.go` (type: `map[string]map[string]depgraph.Mismatch`) and a context for logging, and will be called from there.
  - 2.1.2 Implement a way to update a go dependency in the cloned project using Dagger
  - 2.1.3 Implement a way to commit and push the change to a new branch using Dagger and the same image as 2.1.1

- **2.2. MR Creation Logic**
  - 2.2.1 Implement logic to create merge requests in affected repositories.
    - Status: 
  - 2.2.2 Populate MRs with relevant changes (e.g., dependency version bumps).
    - Status: 

- **2.3. MR Status Tracking**
  - 2.3.1 Track the status of created MRs (open, merged, closed, etc.).
    - Status: 
  - 2.3.2 Notify or log progress for each MR.
    - Status: 

- **2.4. Error Handling & Retries**
  - 2.4.1 Handle failed MR creations or API errors.
    - Status: 
  - 2.4.2 Implement retry and notification mechanisms.
    - Status: 

---

## 3. Integration Coordination

- **3.1. Integration Repository Support**
  - 3.1.1 Identify and support integration repositories (e.g., Docker Compose, Helm Charts).
    - Status: 
  - 3.1.2 Parse and update integration configuration files.
    - Status: 

- **3.2. Final MR Creation**
  - 3.2.1 Create a merge request in the integration repository once all dependent MRs are merged.
    - Status: 
  - 3.2.2 Ensure the integration layer references the latest, compatible versions.
    - Status: 

- **3.3. Consistency & Compatibility Checks**
  - 3.3.1 Validate that all referenced versions are compatible.
    - Status: 
  - 3.3.2 Optionally run integration tests or checks.
    - Status: 

- **3.4. Documentation & Reporting**
  - 3.4.1 Document the update process and results.
    - Status: 
  - 3.4.2 Provide summary reports or dashboards for visibility.
    - Status: 

---

# Development Workflow Rules

## Feature Development Phase
- **First Phase**: Develop the feature, write tests, and ensure code quality
- **Tests and linting MUST ONLY be executed through Dagger** – never run `go test` or `golangci-lint` directly
- Run local tests and linting during development using Dagger commands
- Set subtask status to `in progress` when starting development
- Only proceed to shipping phase when user explicitly approves (e.g., "ship it", "ready to ship", etc.)

_Review and adjust this plan as needed to fit project requirements and priorities._

---

# Completed Tasks

## 1. Dependency Management

- **1.1. Define Configuration Format**
  - 1.1.1 Specify how repositories, services, and libraries are listed in a YAML configuration file.
    - Status: done

- **1.2. Configuration Loading**
  - 1.2.1 Use Viper to create a `main.go` under `cmd/conductor` that loads the YAML config.
    - Status: done

- **1.3. GitHub Client Adapter**
  - 1.3.1 Create an adapter under `internal/adapter` that proxies the GitHub Go client, allowing for interface abstraction and testing.
    - Status: done
  - 1.3.2 Add a test to ensure the adapter can retrieve a file from a repository on GitHub.
    - Status: done
  - 1.3.3 Add a test to ensure the adapter can retrieve the tags of a repository hosted on GitHub.
    - Status: done
  - 1.3.4 Add the tests to a Dagger module (under .dagger), a CI/CD framework, to execute those tests.
    - Status: done

- **1.4. Repository Discovery**
  - 1.4.1 Use the configuration and the adapter to get the content of the configured repositories (no automatic discovery). The business logic should be written in `internal/core`, should use the Github adapter (via interface with dependency injection) and be used by the command in `cmd/conductor`. Use Uber gomock to mock the Github adapter.
    - Status: done

- **1.5. Dependency Graph Construction**
  - 1.5.1 Using the repofetcher, it should pull the go.mod in the repositories listed in the configuration. This should be done in the conductor package.
    - Status: done
  - 1.5.2 Use the dependencies versioning to build a dependency graph: the dependency graph builder should be in its own package under /pkg and be used by /pkg/conductor. The builder must:
    - Receive as input a map of repository module paths to their go.mod file contents (as parsed by the repofetcher)
    - Output a map of module path to a Service struct, where Service contains at least:
      - The module path (string)
      - The repository URL (string)
      - A map of dependencies (map[module path]*Service)
    - There should never be duplicate Service structs for the same module path; all dependencies should point to the same instance
    - The builder should be in its own package under /pkg and be used by /pkg/conductor
    - Tests must be provided for the builder
    - /pkg/conductor should be updated to use the builder
    - Status: done 

- **1.6. Version Detection**
  - 1.6.1 Implement version detection using the GitHub adapter to get tags and detect the last version of each dependency.
    - Status: done
    - Implementation clarifications (2024-06-10):
      - Only semantic versioning tags are considered (e.g., v1.2.3). Non-semver tags are ignored.
      - Pre-release tags (e.g., v1.2.3-beta) are ignored for now.
      - Only repositories/services listed in the config (roots of the dependency graph) are processed.
      - The dependency graph (from depgraph) is the source of the map to update.
      - The Service struct should be updated to include a field for the latest version tag (field name: CurrentVersion).
      - The result is only displayed using a simple print statement (no special formatting).
      - Repositories with no tags are ignored (not an error).
      - Any error fetching tags (other than no tags) should cause the process to fail immediately (fail fast).
      - Only the happy path is tested (no edge cases for now).
      - Tests use only mocks (no real GitHub calls yet) and are placed in a new test file.
      - The logic should be implemented in new files under pkg/repofetcher.
      - Not exposed via CLI or as a public API yet.

## 2. Automated Merge Requests

- **2.1. Repository Authentication & API Integration**
  - 2.1.1 Integrate with GitHub (and/or other platforms) for API access.
    - Status: 
  - 2.1.2 Handle authentication and permissions.
    - Status: 

- **2.2. MR Creation Logic**
  - 2.2.1 Implement logic to create merge requests in affected repositories.
    - Status: 
  - 2.2.2 Populate MRs with relevant changes (e.g., dependency version bumps).
    - Status: 

- **2.3. MR Status Tracking**
  - 2.3.1 Track the status of created MRs (open, merged, closed, etc.).
    - Status: 
  - 2.3.2 Notify or log progress for each MR.
    - Status: 

- **2.4. Error Handling & Retries**
  - 2.4.1 Handle failed MR creations or API errors.
    - Status: 
  - 2.4.2 Implement retry and notification mechanisms.
    - Status: 

## 3. Integration Coordination

- **3.1. Integration Repository Support**
  - 3.1.1 Identify and support integration repositories (e.g., Docker Compose, Helm Charts).
    - Status: 
  - 3.1.2 Parse and update integration configuration files.
    - Status: 

- **3.2. Final MR Creation**
  - 3.2.1 Create a merge request in the integration repository once all dependent MRs are merged.
    - Status: 
  - 3.2.2 Ensure the integration layer references the latest, compatible versions.
    - Status: 

- **3.3. Consistency & Compatibility Checks**
  - 3.3.1 Validate that all referenced versions are compatible.
    - Status: 
  - 3.3.2 Optionally run integration tests or checks.
    - Status: 

- **3.4. Documentation & Reporting**
  - 3.4.1 Document the update process and results.
    - Status: 
  - 3.4.2 Provide summary reports or dashboards for visibility.
    - Status: 

--- 