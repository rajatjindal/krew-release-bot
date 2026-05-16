package releaser

import (
	"fmt"
	"os"
	"time"

	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	ugit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const (
	// OriginNameUpstream is upstream.
	OriginNameUpstream = "upstream"

	// OriginNameLocal is local.
	OriginNameLocal = "local"
)

// CloneRepos clones the repo.
func (r *Releaser) cloneRepos(dir string, request *source.ReleaseRequest) (*ugit.Repository, error) {
	logrus.Infof("Cloning %s", r.UpstreamKrewIndexRepoCloneURL)
	repo, err := ugit.PlainClone(dir, false, &ugit.CloneOptions{
		URL:        r.UpstreamKrewIndexRepoCloneURL,
		Progress:   os.Stdout,
		Auth:       r.getAuth(),
		RemoteName: OriginNameUpstream,
	})
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	r.UpstreamKrewIndexBaseBranch = head.Name().Short()
	logrus.Infof("using upstream base branch %s", r.UpstreamKrewIndexBaseBranch)

	logrus.Infof("Adding remote %s at %s", OriginNameLocal, r.LocalKrewIndexRepoCloneURL)
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: OriginNameLocal,
		URLs: []string{r.LocalKrewIndexRepoCloneURL},
	})
	if err != nil {
		return nil, err
	}

	branchName := r.getBranchName(request)
	logrus.Infof("creating branch %s", branchName)
	err = r.createBranch(repo, branchName)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// CreateBranch creates branch.
func (r *Releaser) createBranch(repo *ugit.Repository, branchName string) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Force:  false,
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err == nil {
		return nil
	}

	return w.Checkout(&git.CheckoutOptions{
		Create: false,
		Force:  false,
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
}

// commitConfig is a git commit.
type commitConfig struct {
	Msg        string
	RemoteName string
}

// AddCommitAndPush commits and pushes.
func (r *Releaser) addCommitAndPush(repo *ugit.Repository, commit commitConfig, request *source.ReleaseRequest) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		return err
	}

	_, err = w.Commit(commit.Msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  r.TokenUsername,
			Email: r.TokenEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	branchName := r.getBranchName(request)
	pushRef := getPushRefSpec(branchName)

	return repo.Push(&ugit.PushOptions{
		RemoteName: commit.RemoteName,
		RefSpecs:   []config.RefSpec{config.RefSpec(pushRef)},
		Auth:       r.getAuth(),
	})
}

func getPushRefSpec(branchName string) string {
	return fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)
}

func (r *Releaser) getAuth() transport.AuthMethod {
	return &githttp.BasicAuth{
		Username: r.TokenUserHandle,
		Password: r.Token,
	}
}
