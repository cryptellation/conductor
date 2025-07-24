package depgraph

// Dependency represents a single dependency of a service.
type Dependency struct {
	Service        *Service
	CurrentVersion string
}

// Service represents a Go module/service in the dependency graph.
type Service struct {
	ModulePath    string
	RepoURL       string
	Dependencies  map[string]Dependency
	LatestVersion string // Latest detected semantic version tag
}
