package depgraph

// Service represents a Go module/service in the dependency graph.
type Service struct {
	ModulePath     string
	RepoURL        string
	Dependencies   map[string]*Service
	CurrentVersion string // Latest detected semantic version tag
}
