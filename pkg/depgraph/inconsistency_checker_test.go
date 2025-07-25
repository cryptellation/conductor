package depgraph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInconsistencyChecker_Check_HappyPath(t *testing.T) {
	// Service A depends on B (outdated) and C (up-to-date)
	serviceB := &Service{
		ModulePath:    "github.com/example/B",
		LatestVersion: "v1.2.0",
	}
	serviceC := &Service{
		ModulePath:    "github.com/example/C",
		LatestVersion: "v2.0.0",
	}
	serviceA := &Service{
		ModulePath: "github.com/example/A",
		Dependencies: map[string]Dependency{
			"github.com/example/B": {Service: serviceB, CurrentVersion: "v1.0.0"}, // outdated
			"github.com/example/C": {Service: serviceC, CurrentVersion: "v2.0.0"}, // up-to-date
		},
	}
	graph := map[string]*Service{
		"github.com/example/A": serviceA,
		"github.com/example/B": serviceB,
		"github.com/example/C": serviceC,
	}

	checker := NewInconsistencyChecker()
	mismatches, err := checker.Check(graph)
	require.NoError(t, err)
	require.Len(t, mismatches, 1)
	deps, ok := mismatches["github.com/example/A"]
	require.True(t, ok)
	require.Len(t, deps, 1)
	mismatch, ok := deps["github.com/example/B"]
	require.True(t, ok)
	require.Equal(t, "v1.0.0", mismatch.Actual)
	require.Equal(t, "v1.2.0", mismatch.Latest)
}
