package releaser

import (
	"testing"

	"github.com/rajatjindal/krew-release-bot/pkg/source"
)

type fakePullRequestOpener struct {
	owner       string
	repo        string
	pullRequest PullRequest
	response    string
}

func (f *fakePullRequestOpener) Open(owner, repo string, pullRequest PullRequest) (string, error) {
	f.owner = owner
	f.repo = repo
	f.pullRequest = pullRequest
	return f.response, nil
}

func TestOpenPullRequestUsesConfiguredOpener(t *testing.T) {
	opener := &fakePullRequestOpener{response: "https://example.com/pr/1"}
	r := &Releaser{
		TokenUserHandle:             "bot-user",
		PullRequestOpener:           opener,
		UpstreamKrewIndexRepoOwner:  "kubernetes-sigs",
		UpstreamKrewIndexRepo:       "krew-index",
		UpstreamKrewIndexBaseBranch: "main",
	}

	request := &source.ReleaseRequest{
		TagName:            "v1.2.3",
		PluginName:         "my-plugin",
		PluginRepo:         "my-plugin",
		PluginOwner:        "acme",
		PluginReleaseActor: "release-user",
	}

	prURL, err := r.openPullRequest(request)
	if err != nil {
		t.Fatalf("openPullRequest returned error: %v", err)
	}

	if prURL != "https://example.com/pr/1" {
		t.Fatalf("unexpected pr url: %s", prURL)
	}

	if opener.owner != "kubernetes-sigs" || opener.repo != "krew-index" {
		t.Fatalf("unexpected target repo: %s/%s", opener.owner, opener.repo)
	}

	if opener.pullRequest.Base != "main" {
		t.Fatalf("unexpected base branch: %s", opener.pullRequest.Base)
	}

	if opener.pullRequest.Head != "bot-user:acme-my-plugin-my-plugin-v1.2.3" {
		t.Fatalf("unexpected head: %s", opener.pullRequest.Head)
	}

	if opener.pullRequest.Title != "release new version v1.2.3 of my-plugin" {
		t.Fatalf("unexpected title: %s", opener.pullRequest.Title)
	}
}

func TestOpenPullRequestRequiresBaseBranch(t *testing.T) {
	r := &Releaser{
		PullRequestOpener:          &fakePullRequestOpener{},
		UpstreamKrewIndexRepoOwner: "kubernetes-sigs",
		UpstreamKrewIndexRepo:      "krew-index",
	}

	_, err := r.openPullRequest(&source.ReleaseRequest{})
	if err == nil || err.Error() != "no upstream base branch configured" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPullRequestUsesLocalBranchNameForDirectMode(t *testing.T) {
	r := &Releaser{
		PRProvider:                  &gitHubPRProvider{},
		UpstreamKrewIndexRepoOwner:  "acme",
		UpstreamKrewIndexRepo:       "custom-index",
		LocalKrewIndexRepoOwner:     "acme",
		LocalKrewIndexRepo:          "custom-index",
		UpstreamKrewIndexBaseBranch: "main",
	}

	pullRequest := r.buildPullRequest(&source.ReleaseRequest{
		TagName:            "v1.2.3",
		PluginName:         "my-plugin",
		PluginRepo:         "my-plugin",
		PluginOwner:        "acme",
		PluginReleaseActor: "release-user",
	})

	if pullRequest.Head != "acme-my-plugin-my-plugin-v1.2.3" {
		t.Fatalf("unexpected direct-mode head: %s", pullRequest.Head)
	}
}

func TestNewRejectsUnsupportedRepoProvider(t *testing.T) {
	_, err := NewWithProviders("stash", ProviderGitHub, "token", IndexRepoConfig{})
	if err == nil || err.Error() != `unsupported git provider "stash"` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRejectsUnsupportedPRProvider(t *testing.T) {
	_, err := NewWithProviders(ProviderGitHub, "stash", "token", IndexRepoConfig{})
	if err == nil || err.Error() != `unsupported pr provider "stash"` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRejectsUnsafeCloneURL(t *testing.T) {
	_, err := New(ProviderGitHub, "token", IndexRepoConfig{
		CloneURL: "https://evil.example/kubernetes-sigs/krew-index.git",
	})
	if err == nil || err.Error() != "github clone url host must be github.com" {
		t.Fatalf("unexpected error: %v", err)
	}
}
