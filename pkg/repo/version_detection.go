//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -destination=mock_version_detection.gen.go -package=repo -source=version_detection.go VersionDetector

package repo

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/cryptellation/conductor/pkg/adapters/github"
	"github.com/cryptellation/conductor/pkg/depgraph"
	gh "github.com/google/go-github/v55/github"
	"golang.org/x/mod/semver"
)

// DetectAndSetCurrentVersions updates the CurrentVersion field of each root service in the dependency graph.
// It fetches tags from GitHub, filters for the latest semantic version (ignoring pre-releases and non-semver),
// and sets the field. Fails fast on any error except for no tags.
func DetectAndSetCurrentVersions(
	ctx context.Context,
	client github.Client,
	services map[string]*depgraph.Service,
) error {
	for _, svc := range services {
		owner, repo := parseOwnerAndRepo(svc.ModulePath)
		if owner == "" || repo == "" {
			return fmt.Errorf("invalid module path: %s", svc.ModulePath)
		}
		tags, err := client.ListTags(ctx, owner, repo)
		if err != nil {
			return fmt.Errorf("error fetching tags for %s: %w", svc.ModulePath, err)
		}
		latest := latestSemverTag(tags)
		if latest != "" {
			svc.LatestVersion = latest
		}
	}
	return nil
}

// latestSemverTag returns the latest semantic version tag (ignoring pre-releases and non-semver tags).
func latestSemverTag(tags []*gh.RepositoryTag) string {
	semverRE := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	var versions []string
	for _, tag := range tags {
		if tag == nil || tag.Name == nil {
			continue
		}
		name := *tag.Name
		if semverRE.MatchString(name) && semver.Prerelease(name) == "" {
			versions = append(versions, name)
		}
	}
	if len(versions) == 0 {
		return ""
	}
	sort.Slice(versions, func(i, j int) bool {
		return semver.Compare(versions[i], versions[j]) > 0 // descending
	})
	return versions[0]
}

// VersionDetector defines the interface for version detection.
type VersionDetector interface {
	DetectAndSetCurrentVersions(ctx context.Context, client github.Client, services map[string]*depgraph.Service) error
}

type versionDetector struct{}

func NewVersionDetector() VersionDetector {
	return &versionDetector{}
}

func (v *versionDetector) DetectAndSetCurrentVersions(
	ctx context.Context,
	client github.Client,
	services map[string]*depgraph.Service,
) error {
	for _, svc := range services {
		owner, repo := parseOwnerAndRepo(svc.ModulePath)
		if owner == "" || repo == "" {
			return fmt.Errorf("invalid module path: %s", svc.ModulePath)
		}
		tags, err := client.ListTags(ctx, owner, repo)
		if err != nil {
			return fmt.Errorf("error fetching tags for %s: %w", svc.ModulePath, err)
		}
		latest := latestSemverTag(tags)
		if latest != "" {
			svc.LatestVersion = latest
		}
	}
	return nil
}
