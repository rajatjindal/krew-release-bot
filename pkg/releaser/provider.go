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

// PullRequestHeadInput captures the fields a PR provider may need to format a head ref.
type PullRequestHeadInput struct {
	BranchName        string
	LocalRepoOwner    string
	LocalRepoName     string
	UpstreamRepoOwner string
	UpstreamRepoName  string
	TokenUserHandle   string
}

// GitProvider encapsulates provider-specific git hosting behavior.
type GitProvider interface {
	Name() string
	DefaultCloneURL(owner, repo string) string
	ResolveCloneURL(owner, repo, override string) (string, error)
	GetAuth(tokenUserHandle, token string) transport.AuthMethod
}

// PRProvider encapsulates provider-specific pull request behavior.
type PRProvider interface {
	Name() string
	NewPullRequestOpener(token string) PullRequestOpener
	FormatPullRequestHead(input PullRequestHeadInput) string
}

type GitProviderFactory func() GitProvider
type PRProviderFactory func() PRProvider

var gitProviders = map[string]GitProviderFactory{}
var prProviders = map[string]PRProviderFactory{}

func init() {
	RegisterGitProvider(ProviderGitHub, func() GitProvider {
		return &gitHubGitProvider{}
	})
	RegisterPRProvider(ProviderGitHub, func() PRProvider {
		return &gitHubPRProvider{}
	})
}

// RegisterGitProvider makes a git hosting provider available for selection.
func RegisterGitProvider(name string, factory GitProviderFactory) {
	normalized := normalizeProviderName(name)
	if normalized == "" || factory == nil {
		return
	}

	gitProviders[normalized] = factory
}

// RegisterPRProvider makes a pull request provider available for selection.
func RegisterPRProvider(name string, factory PRProviderFactory) {
	normalized := normalizeProviderName(name)
	if normalized == "" || factory == nil {
		return
	}

	prProviders[normalized] = factory
}

type gitHubGitProvider struct{}

func (p *gitHubGitProvider) Name() string {
	return ProviderGitHub
}

func (p *gitHubGitProvider) DefaultCloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

func (p *gitHubGitProvider) ResolveCloneURL(owner, repo, override string) (string, error) {
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

func (p *gitHubGitProvider) GetAuth(tokenUserHandle, token string) transport.AuthMethod {
	return &githttp.BasicAuth{
		Username: tokenUserHandle,
		Password: token,
	}
}

type gitHubPRProvider struct{}

func (p *gitHubPRProvider) Name() string {
	return ProviderGitHub
}

func (p *gitHubPRProvider) NewPullRequestOpener(token string) PullRequestOpener {
	return newGitHubPullRequestOpener(token)
}

func (p *gitHubPRProvider) FormatPullRequestHead(input PullRequestHeadInput) string {
	if input.LocalRepoOwner == input.UpstreamRepoOwner && input.LocalRepoName == input.UpstreamRepoName {
		return input.BranchName
	}

	owner := input.LocalRepoOwner
	if owner == "" {
		owner = input.TokenUserHandle
	}

	return fmt.Sprintf("%s:%s", owner, input.BranchName)
}

func getGitProvider(name string) (GitProvider, error) {
	normalized := normalizeProviderName(name)
	if normalized == "" {
		normalized = ProviderGitHub
	}

	factory, ok := gitProviders[normalized]
	if !ok {
		return nil, fmt.Errorf("unsupported git provider %q", name)
	}

	return factory(), nil
}

func getPRProvider(name string) (PRProvider, error) {
	normalized := normalizeProviderName(name)
	if normalized == "" {
		normalized = ProviderGitHub
	}

	factory, ok := prProviders[normalized]
	if !ok {
		return nil, fmt.Errorf("unsupported pr provider %q", name)
	}

	return factory(), nil
}

func normalizeProviderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
