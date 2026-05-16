package releaser

import "github.com/rajatjindal/krew-release-bot/pkg/krew"

// Releaser is what opens PR
type Releaser struct {
	Token                         string
	TokenEmail                    string
	TokenUserHandle               string
	TokenUsername                 string
	RepositoryProvider            RepositoryProvider
	PullRequestOpener             PullRequestOpener
	UpstreamKrewIndexRepo         string
	UpstreamKrewIndexRepoOwner    string
	UpstreamKrewIndexBaseBranch   string
	UpstreamKrewIndexRepoCloneURL string
	LocalKrewIndexRepo            string
	LocalKrewIndexRepoOwner       string
	LocalKrewIndexRepoCloneURL    string
}

// TODO: get email, userhandle, name from token
func getUserDetails(_ string) (string, string, string) {
	return "krew-release-bot", "Krew Release Bot", "krewpluginreleasebot@gmail.com"
}

// New returns new releaser object
func New(providerName, token string) (*Releaser, error) {
	repoProvider, err := getRepositoryProvider(providerName)
	if err != nil {
		return nil, err
	}

	tokenUserHandle, tokenUsername, tokenEmail := getUserDetails(token)

	upstreamCloneURL, err := krew.GetKrewIndexRepoCloneURL(repoProvider, krew.GetKrewIndexRepoOwner(), krew.GetKrewIndexRepoName())
	if err != nil {
		return nil, err
	}

	return &Releaser{
		Token:                         token,
		TokenEmail:                    tokenEmail,
		TokenUserHandle:               tokenUserHandle,
		TokenUsername:                 tokenUsername,
		RepositoryProvider:            repoProvider,
		PullRequestOpener:             repoProvider.NewPullRequestOpener(token),
		UpstreamKrewIndexRepo:         krew.GetKrewIndexRepoName(),
		UpstreamKrewIndexRepoOwner:    krew.GetKrewIndexRepoOwner(),
		UpstreamKrewIndexRepoCloneURL: upstreamCloneURL,
		LocalKrewIndexRepo:            krew.GetKrewIndexRepoName(),
		LocalKrewIndexRepoOwner:       tokenUserHandle,
		LocalKrewIndexRepoCloneURL:    repoProvider.DefaultCloneURL(tokenUserHandle, krew.GetKrewIndexRepoName()),
	}, nil
}

// ConfigureDirectPRs updates the releaser to push branches directly to the target index repo.
func (r *Releaser) ConfigureDirectPRs() {
	r.LocalKrewIndexRepo = r.UpstreamKrewIndexRepo
	r.LocalKrewIndexRepoOwner = r.UpstreamKrewIndexRepoOwner
	r.LocalKrewIndexRepoCloneURL = r.UpstreamKrewIndexRepoCloneURL
}
