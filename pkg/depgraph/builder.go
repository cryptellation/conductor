//go:generate go run go.uber.org/mock/mockgen@v0.2.0 -destination=mock_builder.gen.go -package=depgraph -source=builder.go GraphBuilder
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

// GraphBuilder defines the interface for building dependency graphs.
type GraphBuilder interface {
	BuildGraph(modules map[string]RepoModule) (map[string]*Service, error)
}

type graphBuilder struct{}

func NewGraphBuilder() GraphBuilder {
	return &graphBuilder{}
}

func (g *graphBuilder) BuildGraph(modules map[string]RepoModule) (map[string]*Service, error) {
	// First pass: create all Service nodes (no dependencies yet)
	services := make(map[string]*Service)
	for modulePath := range modules {
		services[modulePath] = &Service{
			ModulePath:    modulePath,
			Dependencies:  make(map[string]Dependency),
			LatestVersion: "",
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
				services[modulePath].Dependencies[depPath] = Dependency{
					Service:        depService,
					CurrentVersion: req.Mod.Version,
				}
			}
			// If dependency is not in the input set, ignore (external dependency)
		}
	}
	return services, nil
}
