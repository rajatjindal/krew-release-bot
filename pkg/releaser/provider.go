package releaser

import "fmt"

const (
	// ProviderGitHub is the currently supported repository/PR provider.
	ProviderGitHub = "github"
)

// RepositoryProvider encapsulates provider-specific repository behavior.
type RepositoryProvider interface {
	Name() string
	CloneURL(owner, repo string) string
	NewPullRequestOpener(token string) PullRequestOpener
}

type gitHubRepositoryProvider struct{}

func (p *gitHubRepositoryProvider) Name() string {
	return ProviderGitHub
}

func (p *gitHubRepositoryProvider) CloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

func (p *gitHubRepositoryProvider) NewPullRequestOpener(token string) PullRequestOpener {
	return newGitHubPullRequestOpener(token)
}

func getRepositoryProvider(name string) (RepositoryProvider, error) {
	switch name {
	case "", ProviderGitHub:
		return &gitHubRepositoryProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported repo/pr provider %q", name)
	}
}
