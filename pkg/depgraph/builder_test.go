//go:build unit
// +build unit

package depgraph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildGraph_Simple(t *testing.T) {
	modA := []byte(`module github.com/example/A
require github.com/example/B v1.0.0
`)
	modB := []byte(`module github.com/example/B
`)
	modules := map[string]RepoModule{
		"github.com/example/A": {RepoURL: "https://github.com/example/A.git", GoModContent: modA},
		"github.com/example/B": {RepoURL: "https://github.com/example/B.git", GoModContent: modB},
	}
	graph, err := NewGraphBuilder().BuildGraph(modules)
	require.NoError(t, err)
	require.Len(t, graph, 2)
	a := graph["github.com/example/A"]
	b := graph["github.com/example/B"]
	require.NotNil(t, a)
	require.NotNil(t, b)
	dep, ok := a.Dependencies["github.com/example/B"]
	require.True(t, ok)
	require.Equal(t, b, dep.Service)
	require.Equal(t, "v1.0.0", dep.CurrentVersion)
}

func TestBuildGraph_SharedDependency(t *testing.T) {
	modA := []byte(`module github.com/example/A
require github.com/example/C/v2 v2.0.0
`)
	modB := []byte(`module github.com/example/B
require github.com/example/C/v2 v2.1.0
`)
	modC := []byte(`module github.com/example/C/v2
`)
	modules := map[string]RepoModule{
		"github.com/example/A":    {RepoURL: "https://github.com/example/A.git", GoModContent: modA},
		"github.com/example/B":    {RepoURL: "https://github.com/example/B.git", GoModContent: modB},
		"github.com/example/C/v2": {RepoURL: "https://github.com/example/C.git", GoModContent: modC},
	}
	graph, err := NewGraphBuilder().BuildGraph(modules)
	require.NoError(t, err)
	a := graph["github.com/example/A"]
	b := graph["github.com/example/B"]
	c := graph["github.com/example/C/v2"]
	require.NotNil(t, a)
	require.NotNil(t, b)
	require.NotNil(t, c)
	depA, okA := a.Dependencies["github.com/example/C/v2"]
	depB, okB := b.Dependencies["github.com/example/C/v2"]
	require.True(t, okA)
	require.True(t, okB)
	require.Equal(t, c, depA.Service)
	require.Equal(t, c, depB.Service)
	require.Equal(t, "v2.0.0", depA.CurrentVersion)
	require.Equal(t, "v2.1.0", depB.CurrentVersion)
	require.True(t, depA.Service == depB.Service)
}

func TestBuildGraph_ExternalDependencyIgnored(t *testing.T) {
	modA := []byte(`module github.com/example/A
require github.com/example/B v1.0.0
require github.com/external/X v1.2.3
`)
	modB := []byte(`module github.com/example/B
`)
	modules := map[string]RepoModule{
		"github.com/example/A": {RepoURL: "https://github.com/example/A.git", GoModContent: modA},
		"github.com/example/B": {RepoURL: "https://github.com/example/B.git", GoModContent: modB},
	}
	graph, err := NewGraphBuilder().BuildGraph(modules)
	require.NoError(t, err)
	a := graph["github.com/example/A"]
	require.NotNil(t, a)
	dep, ok := a.Dependencies["github.com/example/B"]
	require.True(t, ok)
	require.Equal(t, "v1.0.0", dep.CurrentVersion)
	require.NotContains(t, a.Dependencies, "github.com/external/X")
}
