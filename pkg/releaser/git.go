package releaser

import (
	"fmt"
	"os"
	"time"

	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
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

// CloneRepos clones the repo
func (r *Releaser) cloneRepos(dir string, request *source.ReleaseRequest) (*git.Repository, error) {
	logrus.Infof("Cloning %s", r.Config.Upstream.GitCloneURL)
	defaultBranch, err := r.getBaseBranch()
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:           r.Config.Upstream.GitCloneURL,
		Progress:      os.Stdout,
		ReferenceName: getBranchReferenceName(defaultBranch),
		SingleBranch:  true,
		Auth:          r.getAuth(r.Config.Upstream.Auth),
		RemoteName:    OriginNameUpstream,
	})
	if err != nil {
		return nil, err
	}

	logrus.Infof("Adding remote %s at %s", OriginNameLocal, r.Config.LocalPushTarget.GitCloneURL)
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: OriginNameLocal,
		URLs: []string{r.Config.LocalPushTarget.GitCloneURL},
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

// CreateBranch creates branch
func (r *Releaser) createBranch(repo *git.Repository, branchName string) error {
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

// commitConfig is a git commit
type commitConfig struct {
	Msg        string
	RemoteName string
}

// addCommit creates the local commit for the manifest update.
func (r *Releaser) addCommit(repo *git.Repository, commit commitConfig) error {
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

	return nil
}

// pushCommit pushes the created commit to the configured remote branch.
func (r *Releaser) pushCommit(repo *git.Repository, commit commitConfig, request *source.ReleaseRequest) error {
	branchName := r.getBranchName(request)
	pushRef := getPushRefSpec(branchName)

	return repo.Push(&git.PushOptions{
		RemoteName: commit.RemoteName,
		RefSpecs:   []config.RefSpec{config.RefSpec(pushRef)},
		Auth:       r.getAuth(r.Config.LocalPushTarget.Auth),
	})
}

func getPushRefSpec(branchName string) string {
	return fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)
}

func getBranchReferenceName(branch string) plumbing.ReferenceName {
	if plumbing.ReferenceName(branch).IsBranch() {
		return plumbing.ReferenceName(branch)
	}

	return plumbing.NewBranchReferenceName(branch)
}

// SubmitPR submits the PR
func (r *Releaser) submitPR(request *source.ReleaseRequest) (string, error) {
	defaultBranch, err := r.getBaseBranch()
	if err != nil {
		return "", err
	}

	logrus.Infof("creating pr with title %q, \nhead branch %q, \nbase %q, \nbody %q",
		r.getTitle(request),
		r.getBranchName(request),
		defaultBranch,
		r.getPRBody(request),
	)

	pr, err := r.openPullRequest(request, defaultBranch)
	if err != nil {
		return "", err
	}

	logrus.Infof("pr %q opened for releasing new version", pr)
	return pr, nil
}

func (r *Releaser) getTitle(request *source.ReleaseRequest) string {
	return fmt.Sprintf(
		"release new version %s of %s",
		request.TagName,
		request.PluginName,
	)
}

func (r *Releaser) getBranchName(request *source.ReleaseRequest) string {
	return fmt.Sprintf("%s-%s-%s-%s", request.PluginOwner, request.PluginName, request.PluginRepo, request.TagName)
}

func (r *Releaser) getPRBody(request *source.ReleaseRequest) string {
	prBody := `hey krew-index team,

I am [krew-release-bot](https://github.com/rajatjindal/krew-release-bot), and I would like to open this PR to publish version %s of %s on behalf of @%s.

Thanks,
@krew-release-bot`

	return fmt.Sprintf(prBody,
		fmt.Sprintf("`%s`", request.TagName),
		fmt.Sprintf("`%s`", request.PluginName),
		request.PluginReleaseActor,
	)
}

func (r *Releaser) getBaseBranch() (string, error) {
	if r.Config.BaseBranchOverride != "" {
		return r.Config.BaseBranchOverride, nil
	}

	return r.forge.RepoDefaultBranch(r.Config.Upstream.Repo)
}

func (r *Releaser) getAuth(authConfig AuthConfig) transport.AuthMethod {
	return &githttp.BasicAuth{
		Username: r.TokenUserHandle,
		Password: authConfig.Token,
	}
}
