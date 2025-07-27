//go:build unit
// +build unit

package repo

import (
	"context"
	"testing"

	"github.com/cryptellation/conductor/pkg/adapters/github"
	"github.com/cryptellation/conductor/pkg/depgraph"
	gh "github.com/google/go-github/v55/github"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDetectAndSetCurrentVersions_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := github.NewMockClient(ctrl)
	services := map[string]*depgraph.Service{
		"github.com/example/A": {
			ModulePath:   "github.com/example/A",
			Dependencies: map[string]depgraph.Dependency{},
		},
		"github.com/example/B": {
			ModulePath:   "github.com/example/B",
			Dependencies: map[string]depgraph.Dependency{},
		},
	}

	mockClient.EXPECT().ListTags(gomock.Any(), "example", "A").Return([]*gh.RepositoryTag{
		{Name: gh.String("v1.2.3")},
		{Name: gh.String("v1.2.0")},
		{Name: gh.String("v1.2.3-beta")}, // should be ignored
	}, nil)
	mockClient.EXPECT().ListTags(gomock.Any(), "example", "B").Return([]*gh.RepositoryTag{}, nil) // no tags

	err := DetectAndSetCurrentVersions(context.Background(), mockClient, services)
	require.NoError(t, err)
	require.Equal(t, "v1.2.3", services["github.com/example/A"].LatestVersion)
	require.Equal(t, "", services["github.com/example/B"].LatestVersion)
}
