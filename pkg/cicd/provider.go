package cicd

import (
	"fmt"
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
	GetWorkDirectory() string
	GetTemplateFile() string
}

// PreReleaseChecker allows a CI provider to decide whether a release should be skipped.
type PreReleaseChecker interface {
	IsPreRelease(owner, repo, tag string) (bool, error)
}

// Provider groups the CI capabilities used by the application.
type Provider interface {
	ReleaseMetadataProvider
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

var builtInRegistrations = []Registration{
	{
		Name: "github-actions",
		Detect: func() bool {
			return os.Getenv("GITHUB_ACTIONS") == "true"
		},
		New: func() Provider {
			return &github.Actions{}
		},
	},
	{
		Name: "circle-ci",
		Detect: func() bool {
			return os.Getenv("CIRCLECI") == "true"
		},
		New: func() Provider {
			return &circleci.Provider{}
		},
	},
	{
		Name: "travis-ci",
		Detect: func() bool {
			return os.Getenv("TRAVIS") == "true"
		},
		New: func() Provider {
			return &travisci.Provider{}
		},
	},
}

// Registry manages provider registrations explicitly.
type Registry struct {
	providers []Registration
}

// NewRegistry constructs a registry from validated registrations.
func NewRegistry(registrations ...Registration) (*Registry, error) {
	registry := &Registry{}
	for _, reg := range registrations {
		if err := registry.RegisterProvider(reg); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

// RegisterProvider makes a provider available for discovery.
func (r *Registry) RegisterProvider(reg Registration) error {
	if reg.Name == "" {
		return fmt.Errorf("provider registration name is required")
	}
	if reg.Detect == nil {
		return fmt.Errorf("provider registration %q is missing Detect", reg.Name)
	}
	if reg.New == nil {
		return fmt.Errorf("provider registration %q is missing New", reg.Name)
	}

	r.providers = append(r.providers, reg)
	return nil
}

// GetProvider returns the first CI/CD provider whose detection matches the environment.
func (r *Registry) GetProvider() Provider {
	for _, provider := range r.providers {
		if provider.Detect() {
			return provider.New()
		}
	}

	return nil
}

// GetProvider returns the first built-in CI/CD provider whose detection matches the environment.
func GetProvider() (Provider, error) {
	registry, err := NewRegistry(builtInRegistrations...)
	if err != nil {
		return nil, err
	}

	return registry.GetProvider(), nil
}
