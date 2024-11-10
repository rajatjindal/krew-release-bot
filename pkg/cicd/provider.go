package cicd

import (
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/cicd/circleci"
	"github.com/rajatjindal/krew-release-bot/pkg/cicd/github"
	"github.com/rajatjindal/krew-release-bot/pkg/cicd/travisci"
)

// Provider defines CI/CD provider interface
type Provider interface {
	GetTag() (string, error)
	GetActor() (string, error)
	GetOwnerAndRepo() (string, string, error)
	GetWorkDirectory() string
	GetTemplateFile() string
}

// GetProvider returns the CI/CD provider
// e.g. github-actions or circle-ci
func GetProvider() Provider {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return &github.Actions{}
	}

	if os.Getenv("CIRCLECI") == "true" {
		return &circleci.Provider{}
	}

	if os.Getenv("TRAVIS") == "true" {
		return &travisci.Provider{}
	}

	return nil
}
