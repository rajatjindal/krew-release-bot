package releaser

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

type ForgeKind string

const (
	ForgeKindGitHub ForgeKind = "github"
)

type RepoIdentity struct {
	Host          string
	Path          string
	Owner         string
	Name          string
	Metadata      map[string]string
	ProjectID     string
	Namespace     string
	DefaultBranch string
}

func NewRepoIdentity(host, owner, name string) RepoIdentity {
	return RepoIdentity{
		Host:      host,
		Path:      joinRepoPath(owner, name),
		Owner:     owner,
		Name:      name,
		Namespace: owner,
	}
}

func ParseRepoIdentity(forgeKind ForgeKind, repoURL string) (RepoIdentity, error) {
	parsed, err := parseRepoURL(repoURL)
	if err != nil {
		return RepoIdentity{}, err
	}

	repo := RepoIdentity{
		Host:      parsed.Host,
		Path:      parsed.Path,
		Name:      lastPathSegment(parsed.Path),
		Namespace: namespaceFromPath(parsed.Path),
		Metadata: map[string]string{
			"repo_url": repoURL,
		},
	}

	switch forgeKind {
	case "", ForgeKindGitHub, ForgeKind("gitlab"), ForgeKind("bitbucket"):
		repo.Owner = firstPathSegment(parsed.Path)
	default:
		repo.Owner = firstPathSegment(parsed.Path)
	}

	return repo, nil
}

func (r RepoIdentity) FullPath() string {
	if r.Path != "" {
		return r.Path
	}

	return joinRepoPath(r.Owner, r.Name)
}

func (r RepoIdentity) RepoOwner() string {
	if r.Owner != "" {
		return r.Owner
	}

	return firstPathSegment(r.FullPath())
}

func (r RepoIdentity) RepoName() string {
	if r.Name != "" {
		return r.Name
	}

	return lastPathSegment(r.FullPath())
}

func joinRepoPath(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, strings.Trim(part, "/"))
		}
	}

	return strings.Join(filtered, "/")
}

func namespaceFromPath(repoPath string) string {
	parts := strings.Split(strings.Trim(repoPath, "/"), "/")
	if len(parts) <= 1 {
		return ""
	}

	return strings.Join(parts[:len(parts)-1], "/")
}

func firstPathSegment(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

func lastPathSegment(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}

func (r RepoIdentity) ValidateGitHub() error {
	if r.RepoOwner() == "" || r.RepoName() == "" {
		return fmt.Errorf("github repo identity requires owner and name, got path %q", r.FullPath())
	}

	return nil
}

type parsedRepoURL struct {
	Host string
	Path string
}

func parseRepoURL(repoURL string) (parsedRepoURL, error) {
	if strings.HasPrefix(repoURL, "git@") {
		return parseScpLikeRepoURL(repoURL)
	}

	parsed, err := url.Parse(repoURL)
	if err != nil {
		return parsedRepoURL{}, fmt.Errorf("parse repo url %q: %w", repoURL, err)
	}

	if parsed.Host == "" {
		return parsedRepoURL{}, fmt.Errorf("repo url %q is missing host", repoURL)
	}

	repoPath := normalizeRepoPath(parsed.Path)
	if repoPath == "" {
		return parsedRepoURL{}, fmt.Errorf("repo url %q is missing repository path", repoURL)
	}

	return parsedRepoURL{
		Host: parsed.Host,
		Path: repoPath,
	}, nil
}

func parseScpLikeRepoURL(repoURL string) (parsedRepoURL, error) {
	at := strings.Index(repoURL, "@")
	colon := strings.Index(repoURL, ":")
	if at == -1 || colon == -1 || colon < at {
		return parsedRepoURL{}, fmt.Errorf("unsupported ssh repo url %q", repoURL)
	}

	host := repoURL[at+1 : colon]
	repoPath := normalizeRepoPath(repoURL[colon+1:])
	if host == "" || repoPath == "" {
		return parsedRepoURL{}, fmt.Errorf("unsupported ssh repo url %q", repoURL)
	}

	return parsedRepoURL{
		Host: host,
		Path: repoPath,
	}, nil
}

func normalizeRepoPath(raw string) string {
	trimmed := strings.Trim(raw, "/")
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimSuffix(trimmed, ".git")
	cleaned := path.Clean(trimmed)
	if cleaned == "." {
		return ""
	}

	return cleaned
}

type AuthConfig struct {
	Token       string
	TokenEnvVar string
}

type RawReleaseTarget struct {
	ForgeKind  ForgeKind
	APIBaseURL string
	RepoURL    string
	Auth       AuthConfig
	Metadata   map[string]string
}

type RawIndexRepoConfig struct {
	Upstream           RawReleaseTarget
	LocalPushTarget    RawReleaseTarget
	BaseBranchOverride string
	DryRun             bool
}

type ReleaseTarget struct {
	ForgeKind   ForgeKind
	APIBaseURL  string
	GitCloneURL string
	Repo        RepoIdentity
	Auth        AuthConfig
}

type IndexRepoConfig struct {
	Upstream           ReleaseTarget
	LocalPushTarget    ReleaseTarget
	BaseBranchOverride string
	DryRun             bool
}

func IndexRepoConfigFromRaw(raw RawIndexRepoConfig) IndexRepoConfig {
	return IndexRepoConfig{
		Upstream: ReleaseTarget{
			ForgeKind:   raw.Upstream.ForgeKind,
			APIBaseURL:  raw.Upstream.APIBaseURL,
			GitCloneURL: raw.Upstream.RepoURL,
			Repo:        mustParseRepoIdentity(raw.Upstream.ForgeKind, raw.Upstream.RepoURL, raw.Upstream.Metadata),
			Auth:        raw.Upstream.Auth,
		},
		LocalPushTarget: ReleaseTarget{
			ForgeKind:   raw.LocalPushTarget.ForgeKind,
			APIBaseURL:  raw.LocalPushTarget.APIBaseURL,
			GitCloneURL: raw.LocalPushTarget.RepoURL,
			Repo:        mustParseRepoIdentity(raw.LocalPushTarget.ForgeKind, raw.LocalPushTarget.RepoURL, raw.LocalPushTarget.Metadata),
			Auth:        raw.LocalPushTarget.Auth,
		},
		BaseBranchOverride: raw.BaseBranchOverride,
		DryRun:             raw.DryRun,
	}
}

func mustParseRepoIdentity(forgeKind ForgeKind, repoURL string, metadata map[string]string) RepoIdentity {
	repo, err := ParseRepoIdentity(forgeKind, repoURL)
	if err != nil {
		panic(err)
	}

	if metadata != nil {
		if repo.Metadata == nil {
			repo.Metadata = map[string]string{}
		}
		for key, value := range metadata {
			repo.Metadata[key] = value
		}
	}

	return repo
}
