package releaser

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v66/github"
	"github.com/pkg/errors"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"golang.org/x/oauth2"
)

type UserIdentity struct {
	Handle string
	Name   string
	Email  string
}

type PullRequestSpec struct {
	Title      string
	Body       string
	BaseBranch string
	HeadBranch string
	HeadRepo   RepoIdentity
}

type Forge interface {
	CurrentUser() (UserIdentity, error)
	RepoDefaultBranch(repo RepoIdentity) (string, error)
	OpenPullRequest(repo RepoIdentity, spec PullRequestSpec) (string, error)
}

type GitHubForge struct {
	client *github.Client
}

func NewForge(kind ForgeKind, apiBaseURL string, token string) (Forge, error) {
	switch kind {
	case "", ForgeKindGitHub:
		return NewGitHubForge(apiBaseURL, token)
	default:
		return nil, fmt.Errorf("unsupported forge kind %q", kind)
	}
}

func NewGitHubForge(apiBaseURL string, token string) (*GitHubForge, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.TODO(), ts)
	client := github.NewClient(tc)
	if apiBaseURL != "" {
		baseURL, err := url.Parse(apiBaseURL)
		if err != nil {
			return nil, errors.Wrap(err, "invalid github api base url")
		}

		client.BaseURL = baseURL
	}

	return &GitHubForge{client: client}, nil
}

func (g *GitHubForge) CurrentUser() (UserIdentity, error) {
	user, _, err := g.client.Users.Get(context.TODO(), "")
	if err != nil {
		return UserIdentity{}, err
	}

	return UserIdentity{
		Handle: user.GetLogin(),
		Name:   user.GetName(),
		Email:  user.GetEmail(),
	}, nil
}

func (g *GitHubForge) RepoDefaultBranch(repo RepoIdentity) (string, error) {
	repository, _, err := g.client.Repositories.Get(context.TODO(), repo.RepoOwner(), repo.RepoName())
	if err != nil {
		return "", err
	}

	return repository.GetDefaultBranch(), nil
}

func (g *GitHubForge) OpenPullRequest(repo RepoIdentity, spec PullRequestSpec) (string, error) {
	head := fmt.Sprintf("%s:%s", spec.HeadRepo.RepoOwner(), spec.HeadBranch)
	pr, _, err := g.client.PullRequests.Create(
		context.TODO(),
		repo.RepoOwner(),
		repo.RepoName(),
		&github.NewPullRequest{
			Title: github.String(spec.Title),
			Head:  github.String(head),
			Base:  github.String(spec.BaseBranch),
			Body:  github.String(spec.Body),
		},
	)
	if err != nil {
		return "", err
	}

	return pr.GetHTMLURL(), nil
}

func (r *Releaser) openPullRequest(request *source.ReleaseRequest, base string) (string, error) {
	return r.forge.OpenPullRequest(r.Config.Upstream.Repo, PullRequestSpec{
		Title:      r.getTitle(request),
		Body:       r.getPRBody(request),
		BaseBranch: base,
		HeadBranch: r.getBranchName(request),
		HeadRepo:   r.Config.LocalPushTarget.Repo,
	})
}
