package releaser

import (
	"fmt"

	"github.com/rajatjindal/krew-release-bot/pkg/krew"
)

// Releaser is what opens PR
type Releaser struct {
	Token                         string
	TokenEmail                    string
	TokenUserHandle               string
	TokenUsername                 string
	PullRequestOpener             PullRequestOpener
	UpstreamKrewIndexRepo         string
	UpstreamKrewIndexRepoOwner    string
	UpstreamKrewIndexBaseBranch   string
	UpstreamKrewIndexRepoCloneURL string
	LocalKrewIndexRepo            string
	LocalKrewIndexRepoOwner       string
	LocalKrewIndexRepoCloneURL    string
}

func getCloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

// TODO: get email, userhandle, name from token
func getUserDetails(_ string) (string, string, string) {
	return "krew-release-bot", "Krew Release Bot", "krewpluginreleasebot@gmail.com"
}

// New returns new releaser object
func New(ghToken string) *Releaser {
	tokenUserHandle, tokenUsername, tokenEmail := getUserDetails(ghToken)

	return &Releaser{
		Token:                         ghToken,
		TokenEmail:                    tokenEmail,
		TokenUserHandle:               tokenUserHandle,
		TokenUsername:                 tokenUsername,
		PullRequestOpener:             newGitHubPullRequestOpener(ghToken),
		UpstreamKrewIndexRepo:         krew.GetKrewIndexRepoName(),
		UpstreamKrewIndexRepoOwner:    krew.GetKrewIndexRepoOwner(),
		UpstreamKrewIndexRepoCloneURL: getCloneURL(krew.GetKrewIndexRepoOwner(), krew.GetKrewIndexRepoName()),
		LocalKrewIndexRepo:            krew.GetKrewIndexRepoName(),
		LocalKrewIndexRepoOwner:       tokenUserHandle,
		LocalKrewIndexRepoCloneURL:    "https://github.com/krew-release-bot/krew-index.git",
	}
}

// ConfigureDirectPRs updates the releaser to push branches directly to the target index repo.
func (r *Releaser) ConfigureDirectPRs() {
	r.LocalKrewIndexRepo = r.UpstreamKrewIndexRepo
	r.LocalKrewIndexRepoOwner = r.UpstreamKrewIndexRepoOwner
	r.LocalKrewIndexRepoCloneURL = r.UpstreamKrewIndexRepoCloneURL
}
