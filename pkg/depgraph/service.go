package depgraph

// Dependency represents a single dependency of a service.
type Dependency struct {
	Service        *Service
	CurrentVersion string
}

// Service represents a Go module/service in the dependency graph.
type Service struct {
	ModulePath    string
	Dependencies  map[string]Dependency
	LatestVersion string // Latest detected semantic version tag
}

// Mismatch represents a version inconsistency between the actual and latest version of a dependency.
type Mismatch struct {
	Actual string
	Latest string
}
