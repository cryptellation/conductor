package depgraph

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// InconsistencyChecker checks for version mismatches between used and latest dependency versions in a dependency graph.
//
//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -destination=mock_inconsistency_checker.gen.go -package=depgraph . InconsistencyChecker
type InconsistencyChecker interface {
	// Check inspects the dependency graph and returns a map of service module path to dependency module path to Mismatch.
	// Only mismatches are included in the result.
	Check(graph map[string]*Service) (map[string]map[string]Mismatch, error)
}

// inconsistencyChecker is the default implementation of InconsistencyChecker.
type inconsistencyChecker struct{}

// NewInconsistencyChecker creates a new InconsistencyChecker.
func NewInconsistencyChecker() InconsistencyChecker {
	return &inconsistencyChecker{}
}

// Check implements the InconsistencyChecker interface.
func (c *inconsistencyChecker) Check(graph map[string]*Service) (map[string]map[string]Mismatch, error) {
	result := make(map[string]map[string]Mismatch)
	for svcPath, svc := range graph {
		if svc == nil {
			continue
		}
		for depPath, dep := range svc.Dependencies {
			if dep.Service == nil {
				continue
			}
			// Skip if no latest version detected
			if dep.Service.LatestVersion == "" {
				continue
			}
			// Parse versions
			actualVer, err := semver.NewVersion(dep.CurrentVersion)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse actual version '%s' for dependency '%s' in service '%s': %w",
					dep.CurrentVersion, depPath, svcPath, err,
				)
			}
			latestVer, err := semver.NewVersion(dep.Service.LatestVersion)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse latest version '%s' for dependency '%s': %w",
					dep.Service.LatestVersion, depPath, err,
				)
			}
			if actualVer.LessThan(latestVer) {
				if result[svcPath] == nil {
					result[svcPath] = make(map[string]Mismatch)
				}
				result[svcPath][depPath] = Mismatch{
					Actual: dep.CurrentVersion,
					Latest: dep.Service.LatestVersion,
				}
			}
		}
	}
	return result, nil
}
