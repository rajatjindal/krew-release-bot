package provider

import (
	"os"

	githubactions "github.com/rajatjindal/krew-release-bot/pkg/provider/github"
)

//Provider defines provider interface
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
		return &githubactions.Provider{}
	}

	return nil
}
