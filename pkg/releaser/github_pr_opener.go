package releaser

import (
	"context"

	"github.com/google/go-github/v66/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// PullRequest describes the data needed by a PR backend.
type PullRequest struct {
	Title string
	Head  string
	Base  string
	Body  string
}

// PullRequestOpener opens pull requests against a remote provider.
type PullRequestOpener interface {
	Open(owner, repo string, pullRequest PullRequest) (string, error)
}

type githubPullRequestOpener struct {
	token string
}

func newGitHubPullRequestOpener(token string) PullRequestOpener {
	return &githubPullRequestOpener{token: token}
}

func (o *githubPullRequestOpener) Open(owner, repo string, pullRequest PullRequest) (string, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: o.token})
	tc := oauth2.NewClient(context.TODO(), ts)
	client := github.NewClient(tc)

	prr := &github.NewPullRequest{
		Title: github.String(pullRequest.Title),
		Head:  github.String(pullRequest.Head),
		Base:  github.String(pullRequest.Base),
		Body:  github.String(pullRequest.Body),
	}

	logrus.Infof("creating pr with title %q, \nhead %q, \nbase %q, \nbody %q",
		github.Stringify(prr.Title),
		github.Stringify(prr.Head),
		github.Stringify(prr.Base),
		github.Stringify(prr.Body),
	)

	pr, _, err := client.PullRequests.Create(context.TODO(), owner, repo, prr)
	if err != nil {
		return "", err
	}

	logrus.Infof("pr %q opened for releasing new version", pr.GetHTMLURL())
	return pr.GetHTMLURL(), nil
}
