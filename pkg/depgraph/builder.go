package depgraph

import (
	"fmt"

	"golang.org/x/mod/modfile"
)

// RepoModule represents the input for the builder: a repo URL and its go.mod content.
type RepoModule struct {
	RepoURL      string
	GoModContent []byte
}

// BuildGraph builds the dependency graph from the given modules.
// Input: map[modulePath]RepoModule
// Output: map[modulePath]*Service.
func BuildGraph(modules map[string]RepoModule) (map[string]*Service, error) {
	// First pass: create all Service nodes (no dependencies yet)
	services := make(map[string]*Service)
	for modulePath, repo := range modules {
		services[modulePath] = &Service{
			ModulePath:   modulePath,
			RepoURL:      repo.RepoURL,
			Dependencies: make(map[string]*Service),
		}
	}

	// Second pass: parse go.mod and wire dependencies
	for modulePath, repo := range modules {
		mf, err := modfile.Parse(fmt.Sprintf("%s/go.mod", modulePath), repo.GoModContent, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to parse go.mod for %s: %w", modulePath, err)
		}
		for _, req := range mf.Require {
			depPath := req.Mod.Path
			if depService, ok := services[depPath]; ok {
				services[modulePath].Dependencies[depPath] = depService
			}
			// If dependency is not in the input set, ignore (external dependency)
		}
	}
	return services, nil
}
