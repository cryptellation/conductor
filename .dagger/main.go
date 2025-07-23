// A generated module for Conductor functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"dagger/conductor/internal/dagger"
)

type Conductor struct{}

// IntegrationTests runs all Go tests in internal/adapters/.
func (m *Conductor) IntegrationTests(sourceDir *dagger.Directory, githubToken *dagger.Secret) *dagger.Container {
	c := dag.Container().From("golang:1.24")
	c = c.WithMountedDirectory("/src", sourceDir).
		WithWorkdir("/src")
	c = c.WithSecretVariable("GITHUB_TOKEN", githubToken)
	return c.WithExec([]string{"go", "test", "./internal/adapters/...", "-v"})
}

// Lint runs golangci-lint on the main repo (./...) only.
func (m *Conductor) Lint(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().
		From("golangci/golangci-lint:v1.62.0").
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint"))

	c = c.WithMountedDirectory("/src", sourceDir).
		WithWorkdir("/src")

	// Lint main repo only
	c = c.WithExec([]string{"golangci-lint", "run", "--timeout", "10m", "./..."})

	return c
}

// LintDagger runs golangci-lint on the .dagger directory only.
func (m *Conductor) LintDagger(sourceDir *dagger.Directory) *dagger.Container {
	c := dag.Container().
		From("golangci/golangci-lint:v1.62.0").
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint"))

	c = c.WithMountedDirectory("/src", sourceDir).
		WithWorkdir("/src")

	// Lint .dagger directory using parent config and module context
	c = c.WithExec([]string{"sh", "-c", "cd .dagger && golangci-lint run --config ../.golangci.yml --timeout 10m ."})

	return c
}
