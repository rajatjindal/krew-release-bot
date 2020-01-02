package releaser

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	ugit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const (
	//OriginNameUpstream is upstream
	OriginNameUpstream = "upstream"

	//OriginNameLocal is local
	OriginNameLocal = "local"
)

//CloneRepos clones the repo
func (r *Releaser) cloneRepos(dir string, request *source.ReleaseRequest) (*ugit.Repository, error) {
	logrus.Infof("Cloning %s", r.UpstreamKrewIndexRepoCloneURL)
	repo, err := ugit.PlainClone(dir, false, &ugit.CloneOptions{
		URL:           r.UpstreamKrewIndexRepoCloneURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.Master,
		SingleBranch:  true,
		Auth:          r.getAuth(),
		RemoteName:    OriginNameUpstream,
	})
	if err != nil {
		return nil, err
	}

	logrus.Infof("Adding remote %s at %s", OriginNameLocal, r.LocalKrewIndexRepoCloneURL)
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: OriginNameLocal,
		URLs: []string{r.LocalKrewIndexRepoCloneURL},
	})
	if err != nil {
		return nil, err
	}

	branchName := r.getBranchName(request)
	logrus.Infof("creating branch %s", *branchName)
	err = r.createBranch(repo, *branchName)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

//CreateBranch creates branch
func (r *Releaser) createBranch(repo *ugit.Repository, branchName string) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// First try to create branch
	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Force:  false,
		Branch: plumbing.NewBranchReferenceName(branchName),
	})

	if err == nil {
		return nil
	}

	//may be it already exists
	return w.Checkout(&git.CheckoutOptions{
		Create: false,
		Force:  false,
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
}

//commitConfig is a git commit
type commitConfig struct {
	Msg        string
	RemoteName string
}

//AddCommitAndPush commits and push
func (r *Releaser) addCommitAndPush(repo *ugit.Repository, commit commitConfig, request *source.ReleaseRequest) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	w.Add(".")
	_, err = w.Commit(commit.Msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  r.TokenUsername,
			Email: r.TokenEmail,
			When:  time.Now(),
		},
	})

	return repo.Push(&ugit.PushOptions{
		RemoteName: commit.RemoteName,
		RefSpecs:   []config.RefSpec{config.DefaultPushRefSpec},
		Auth:       r.getAuth(),
	})
}

//SubmitPR submits the PR
func (r *Releaser) submitPR(request *source.ReleaseRequest) (string, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: r.Token})
	tc := oauth2.NewClient(context.TODO(), ts)
	client := github.NewClient(tc)

	prr := &github.NewPullRequest{
		Title: r.getTitle(request),
		Head:  r.getHead(request),
		Base:  github.String("master"),
		Body:  r.getPRBody(request),
	}

	logrus.Infof("creating pr with title %q, \nhead %q, \nbase %q, \nbody %q",
		github.Stringify(r.getTitle(request)),
		github.Stringify(r.getHead(request)),
		"master",
		github.Stringify(r.getPRBody(request)),
	)

	pr, _, err := client.PullRequests.Create(
		context.TODO(),
		r.UpstreamKrewIndexRepoOwner,
		r.UpstreamKrewIndexRepo,
		prr,
	)
	if err != nil {
		return "", err
	}

	logrus.Infof("pr %q opened for releasing new version", pr.GetHTMLURL())
	return pr.GetHTMLURL(), nil
}

func (r *Releaser) getTitle(request *source.ReleaseRequest) *string {
	s := fmt.Sprintf(
		"release new version %s of %s",
		request.TagName,
		request.PluginName,
	)

	return github.String(s)
}

func (r *Releaser) getBranchName(request *source.ReleaseRequest) *string {
	s := fmt.Sprintf("%s-%s-%s", request.PluginOwner, request.PluginRepo, request.TagName)
	fmt.Printf("creating branch %s", s)
	return github.String(s)
}

func (r *Releaser) getHead(request *source.ReleaseRequest) *string {
	branchName := r.getBranchName(request)
	s := fmt.Sprintf("%s:%s", r.TokenUserHandle, *branchName)
	return github.String(s)
}

func (r *Releaser) getPRBody(request *source.ReleaseRequest) *string {
	prBody := `hey krew-index team,
I am [krew-release-bot](https://github.com/rajatjindal/krew-release-bot), and I would like to open this PR to publish version %s of %s on behalf of [%s](https://github.com/%s).
Thanks,
[krew-release-bot](https://github.com/rajatjindal/krew-release-bot)`

	s := fmt.Sprintf(prBody,
		fmt.Sprintf("`%s`", request.TagName),
		fmt.Sprintf("`%s`", request.PluginName),
		request.PluginReleaseActor,
		request.PluginReleaseActor,
	)

	return github.String(s)
}

func (r *Releaser) getAuth() transport.AuthMethod {
	return &githttp.BasicAuth{
		Username: r.TokenUserHandle,
		Password: r.Token,
	}
}
