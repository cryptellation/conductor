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
	graph, err := BuildGraph(modules)
	require.NoError(t, err)
	require.Len(t, graph, 2)
	a := graph["github.com/example/A"]
	b := graph["github.com/example/B"]
	require.NotNil(t, a)
	require.NotNil(t, b)
	require.Equal(t, b, a.Dependencies["github.com/example/B"])
}

func TestBuildGraph_SharedDependency(t *testing.T) {
	modA := []byte(`module github.com/example/A
require github.com/example/C v1.0.0
`)
	modB := []byte(`module github.com/example/B
require github.com/example/C v1.0.0
`)
	modC := []byte(`module github.com/example/C
`)
	modules := map[string]RepoModule{
		"github.com/example/A": {RepoURL: "https://github.com/example/A.git", GoModContent: modA},
		"github.com/example/B": {RepoURL: "https://github.com/example/B.git", GoModContent: modB},
		"github.com/example/C": {RepoURL: "https://github.com/example/C.git", GoModContent: modC},
	}
	graph, err := BuildGraph(modules)
	require.NoError(t, err)
	a := graph["github.com/example/A"]
	b := graph["github.com/example/B"]
	c := graph["github.com/example/C"]
	require.NotNil(t, a)
	require.NotNil(t, b)
	require.NotNil(t, c)
	require.Equal(t, c, a.Dependencies["github.com/example/C"])
	require.Equal(t, c, b.Dependencies["github.com/example/C"])
	require.True(t, a.Dependencies["github.com/example/C"] == b.Dependencies["github.com/example/C"])
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
	graph, err := BuildGraph(modules)
	require.NoError(t, err)
	a := graph["github.com/example/A"]
	require.NotNil(t, a)
	require.Contains(t, a.Dependencies, "github.com/example/B")
	require.NotContains(t, a.Dependencies, "github.com/external/X")
}
