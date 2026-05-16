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

type GitProviderRegistration struct {
	Name    string
	Factory GitProviderFactory
}

type PRProviderRegistration struct {
	Name    string
	Factory PRProviderFactory
}

var builtInGitProviderRegistrations = []GitProviderRegistration{
	{
		Name: ProviderGitHub,
		Factory: func() GitProvider {
			return &gitHubGitProvider{}
		},
	},
}

var builtInPRProviderRegistrations = []PRProviderRegistration{
	{
		Name: ProviderGitHub,
		Factory: func() PRProvider {
			return &gitHubPRProvider{}
		},
	},
}

type GitProviderRegistry struct {
	providers map[string]GitProviderFactory
}

type PRProviderRegistry struct {
	providers map[string]PRProviderFactory
}

func NewGitProviderRegistry(registrations ...GitProviderRegistration) (*GitProviderRegistry, error) {
	registry := &GitProviderRegistry{
		providers: map[string]GitProviderFactory{},
	}

	for _, reg := range registrations {
		if err := registry.RegisterProvider(reg); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

func NewPRProviderRegistry(registrations ...PRProviderRegistration) (*PRProviderRegistry, error) {
	registry := &PRProviderRegistry{
		providers: map[string]PRProviderFactory{},
	}

	for _, reg := range registrations {
		if err := registry.RegisterProvider(reg); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

// RegisterProvider makes a git hosting provider available for selection.
func (r *GitProviderRegistry) RegisterProvider(reg GitProviderRegistration) error {
	normalized := normalizeProviderName(reg.Name)
	if normalized == "" {
		return fmt.Errorf("git provider registration name is required")
	}
	if reg.Factory == nil {
		return fmt.Errorf("git provider registration %q is missing Factory", reg.Name)
	}

	r.providers[normalized] = reg.Factory
	return nil
}

// RegisterProvider makes a pull request provider available for selection.
func (r *PRProviderRegistry) RegisterProvider(reg PRProviderRegistration) error {
	normalized := normalizeProviderName(reg.Name)
	if normalized == "" {
		return fmt.Errorf("pr provider registration name is required")
	}
	if reg.Factory == nil {
		return fmt.Errorf("pr provider registration %q is missing Factory", reg.Name)
	}

	r.providers[normalized] = reg.Factory
	return nil
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

func (r *GitProviderRegistry) GetProvider(name string) (GitProvider, error) {
	normalized := normalizeProviderName(name)
	if normalized == "" {
		normalized = ProviderGitHub
	}

	factory, ok := r.providers[normalized]
	if !ok {
		return nil, fmt.Errorf("unsupported git provider %q", name)
	}

	return factory(), nil
}

func (r *PRProviderRegistry) GetProvider(name string) (PRProvider, error) {
	normalized := normalizeProviderName(name)
	if normalized == "" {
		normalized = ProviderGitHub
	}

	factory, ok := r.providers[normalized]
	if !ok {
		return nil, fmt.Errorf("unsupported pr provider %q", name)
	}

	return factory(), nil
}

func getGitProvider(name string) (GitProvider, error) {
	registry, err := NewGitProviderRegistry(builtInGitProviderRegistrations...)
	if err != nil {
		return nil, err
	}

	return registry.GetProvider(name)
}

func getPRProvider(name string) (PRProvider, error) {
	registry, err := NewPRProviderRegistry(builtInPRProviderRegistrations...)
	if err != nil {
		return nil, err
	}

	return registry.GetProvider(name)
}

func normalizeProviderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
