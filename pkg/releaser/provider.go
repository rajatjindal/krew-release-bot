package releaser

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const (
	// ProviderGitHub is the currently supported repository/PR provider.
	ProviderGitHub = "github"
)

// PullRequestHeadInput captures the fields a provider may need to format a PR head ref.
type PullRequestHeadInput struct {
	BranchName        string
	LocalRepoOwner    string
	LocalRepoName     string
	UpstreamRepoOwner string
	UpstreamRepoName  string
	TokenUserHandle   string
}

// RepositoryProvider encapsulates provider-specific repository behavior.
type RepositoryProvider interface {
	Name() string
	DefaultCloneURL(owner, repo string) string
	ResolveCloneURL(owner, repo, override string) (string, error)
	NewPullRequestOpener(token string) PullRequestOpener
	GetAuth(tokenUserHandle, token string) transport.AuthMethod
	FormatPullRequestHead(input PullRequestHeadInput) string
}

type RepositoryProviderFactory func() RepositoryProvider

var repositoryProviders = map[string]RepositoryProviderFactory{}

func init() {
	RegisterRepositoryProvider(ProviderGitHub, func() RepositoryProvider {
		return &gitHubRepositoryProvider{}
	})
}

// RegisterRepositoryProvider makes a repo/PR provider available for selection.
func RegisterRepositoryProvider(name string, factory RepositoryProviderFactory) {
	if name == "" || factory == nil {
		return
	}

	repositoryProviders[name] = factory
}

type gitHubRepositoryProvider struct{}

func (p *gitHubRepositoryProvider) Name() string {
	return ProviderGitHub
}

func (p *gitHubRepositoryProvider) DefaultCloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

func (p *gitHubRepositoryProvider) ResolveCloneURL(owner, repo, override string) (string, error) {
	if override == "" {
		return p.DefaultCloneURL(owner, repo), nil
	}

	parsed, err := url.Parse(override)
	if err != nil {
		return "", fmt.Errorf("invalid clone url %q: %w", override, err)
	}

	if parsed.Scheme != "https" {
		return "", fmt.Errorf("github clone url must use https")
	}

	if parsed.Host != "github.com" {
		return "", fmt.Errorf("github clone url host must be github.com")
	}

	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("github clone url must not include credentials, query params, or fragments")
	}

	expectedPath := path.Clean("/" + owner + "/" + repo + ".git")
	if path.Clean(parsed.Path) != expectedPath {
		return "", fmt.Errorf("github clone url path must match /%s/%s.git", owner, repo)
	}

	return override, nil
}

func (p *gitHubRepositoryProvider) NewPullRequestOpener(token string) PullRequestOpener {
	return newGitHubPullRequestOpener(token)
}

func (p *gitHubRepositoryProvider) GetAuth(tokenUserHandle, token string) transport.AuthMethod {
	return &githttp.BasicAuth{
		Username: tokenUserHandle,
		Password: token,
	}
}

func (p *gitHubRepositoryProvider) FormatPullRequestHead(input PullRequestHeadInput) string {
	if input.LocalRepoOwner == input.UpstreamRepoOwner && input.LocalRepoName == input.UpstreamRepoName {
		return input.BranchName
	}

	owner := input.LocalRepoOwner
	if owner == "" {
		owner = input.TokenUserHandle
	}

	return fmt.Sprintf("%s:%s", owner, input.BranchName)
}

func getRepositoryProvider(name string) (RepositoryProvider, error) {
	if name == "" {
		name = ProviderGitHub
	}

	factory, ok := repositoryProviders[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unsupported repo/pr provider %q", name)
	}

	return factory(), nil
}
