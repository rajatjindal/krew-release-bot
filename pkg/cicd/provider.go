package cicd

import (
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/cicd/circleci"
	"github.com/rajatjindal/krew-release-bot/pkg/cicd/github"
	"github.com/rajatjindal/krew-release-bot/pkg/cicd/travisci"
)

// ReleaseMetadataProvider exposes the CI data needed to build a release request.
type ReleaseMetadataProvider interface {
	GetTag() (string, error)
	GetActor() (string, error)
	GetOwnerAndRepo() (string, string, error)
	GetTemplateFile() string
}

// WorkDirectoryProvider exposes the CI checkout directory when needed.
type WorkDirectoryProvider interface {
	GetWorkDirectory() string
}

// PreReleaseChecker allows a CI provider to decide whether a release should be skipped.
type PreReleaseChecker interface {
	IsPreRelease(owner, repo, tag string) (bool, error)
}

// Provider groups the CI capabilities used by the application.
type Provider interface {
	ReleaseMetadataProvider
	WorkDirectoryProvider
	PreReleaseChecker
}

// Factory constructs a CI provider implementation.
type Factory func() Provider

// Registration describes how a provider is detected and constructed.
type Registration struct {
	Name   string
	Detect func() bool
	New    Factory
}

var providers []Registration

func init() {
	RegisterProvider(Registration{
		Name: "github-actions",
		Detect: func() bool {
			return os.Getenv("GITHUB_ACTIONS") == "true"
		},
		New: func() Provider {
			return &github.Actions{}
		},
	})

	RegisterProvider(Registration{
		Name: "circle-ci",
		Detect: func() bool {
			return os.Getenv("CIRCLECI") == "true"
		},
		New: func() Provider {
			return &circleci.Provider{}
		},
	})

	RegisterProvider(Registration{
		Name: "travis-ci",
		Detect: func() bool {
			return os.Getenv("TRAVIS") == "true"
		},
		New: func() Provider {
			return &travisci.Provider{}
		},
	})
}

// RegisterProvider makes a provider available for discovery.
func RegisterProvider(reg Registration) {
	if reg.Detect == nil || reg.New == nil {
		return
	}

	providers = append(providers, reg)
}

// GetProvider returns the first CI/CD provider whose detection matches the environment.
func GetProvider() Provider {
	for _, provider := range providers {
		if provider.Detect() {
			return provider.New()
		}
	}

	return nil
}
