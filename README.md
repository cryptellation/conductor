# Depsync

Golang dependencies synchronizer for polyrepository projects on GitHub

## Overview

Depsync is a tool designed to manage and coordinate dependencies between multiple services and libraries in a polyrepository (multi-repo) architecture. It automates the process of tracking, updating, and integrating dependencies across various repositories, ensuring that all services and libraries remain compatible and up-to-date.

## Key Features

- **Dependency Management:**
  - Analyzes all specified libraries and services to map out their dependencies.
  - Detects new versions of libraries or services and identifies which other services or libraries depend on them.

- **Automated Merge Requests:**
  - Automatically creates merge requests (MRs) in repositories that are affected by dependency updates.
  - Tracks the status of these MRs to ensure all dependent services are updated accordingly.

- **Integration Coordination:**
  - Once all dependent services and libraries have been updated and their MRs have been merged, DepSync can create a final MR in the integration repository (such as a Docker Compose or Helm Chart repository).
  - This ensures that the integration layer always references the latest, compatible versions of all services and libraries, reducing the risk of integration issues.

## Workflow

1. **Dependency Analysis:**
   - Depsync scans the provided list of libraries and services, building a dependency graph.
   - It identifies which services depend on which libraries, and which libraries or services are common dependencies.

2. **Version Detection:**
   - The tool checks for new versions of libraries and services.
   - When a new version is detected, it determines which downstream services or libraries are affected.

3. **Automated Updates:**
   - For each affected repository, Depsync creates a merge request to update the dependency version.
   - It monitors the status of these MRs, ensuring that all required updates are merged.

4. **Integration Update:**
   - Once all dependent repositories have been updated and merged, Depsync creates a merge request in the integration repository (e.g., Docker Compose or Helm Chart repo).
   - This final MR ensures that the deployment configuration references the latest versions of all components.

## Use Cases

- Keeping microservices and shared libraries in sync across multiple repositories.
- Automating the propagation of dependency updates through a service mesh.
- Ensuring integration repositories (Docker Compose, Helm Charts) always use compatible versions of all services.

## Dependencies

- [spf13/viper](https://github.com/spf13/viper) - for configuration loading

## Getting Started

_Coming soon: Instructions on installation, configuration, and usage._

## License

This project is licensed under the GNU General Public License v3.0. See the [LICENSE](LICENSE) file for details.
