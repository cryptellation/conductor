# Development Plan for Conductor

This document outlines the development plan for the Conductor tool, broken down into main tasks with distinct subtasks for each phase. Each subtask includes a line for tracking status.

---

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
  - 1.5.1 Using the repofetcher, it should pull the go.mod in the repositories listed
  in the configuration. This should be done in the conductor package.
    - Status: done
  - 1.5.2 Use the dependencies versioning to build a dependency graph.
    - Status: to do

- **1.6. Version Detection**
  - 1.6.1 Implement version detection using the GitHub adapter to get tags and detect the last version of each dependency.
    - Status: to do
  - 1.6.2 Use the GitHub adapter to get files (e.g., go.mod) and detect the current version of dependencies used in each service or library.
    - Status: to do

- **1.7. Inconsistency Detection**
  - 1.7.1 Detect inconsistencies between the last version of a dependency and the version used in libraries/services on all repositories, based on the dependency graph and version detection.
    - Status: to do

---

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

_Review and adjust this plan as needed to fit project requirements and priorities._ 