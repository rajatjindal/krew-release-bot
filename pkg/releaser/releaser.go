package releaser

import (
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/krew"
)

// Releaser is what opens PR
type Releaser struct {
	Token                         string
	TokenEmail                    string
	TokenUserHandle               string
	TokenUsername                 string
	GitProvider                   GitProvider
	PRProvider                    PRProvider
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
	return NewWithProviders(providerName, providerName, token)
}

// NewWithProviders returns a releaser configured with separate git and PR providers.
func NewWithProviders(gitProviderName, prProviderName, token string) (*Releaser, error) {
	gitProvider, err := getGitProvider(gitProviderName)
	if err != nil {
		return nil, err
	}
	prProvider, err := getPRProvider(prProviderName)
	if err != nil {
		return nil, err
	}

	tokenUserHandle, tokenUsername, tokenEmail := getUserDetails(token)

	upstreamCloneURL, err := getUpstreamKrewIndexRepoCloneURL(gitProvider, krew.GetKrewIndexRepoOwner(), krew.GetKrewIndexRepoName())
	if err != nil {
		return nil, err
	}

	return &Releaser{
		Token:                         token,
		TokenEmail:                    tokenEmail,
		TokenUserHandle:               tokenUserHandle,
		TokenUsername:                 tokenUsername,
		GitProvider:                   gitProvider,
		PRProvider:                    prProvider,
		PullRequestOpener:             prProvider.NewPullRequestOpener(token),
		UpstreamKrewIndexRepo:         krew.GetKrewIndexRepoName(),
		UpstreamKrewIndexRepoOwner:    krew.GetKrewIndexRepoOwner(),
		UpstreamKrewIndexRepoCloneURL: upstreamCloneURL,
		LocalKrewIndexRepo:            krew.GetKrewIndexRepoName(),
		LocalKrewIndexRepoOwner:       tokenUserHandle,
		LocalKrewIndexRepoCloneURL:    gitProvider.DefaultCloneURL(tokenUserHandle, krew.GetKrewIndexRepoName()),
	}, nil
}

// ConfigureDirectPRs updates the releaser to push branches directly to the target index repo.
func (r *Releaser) ConfigureDirectPRs() {
	r.LocalKrewIndexRepo = r.UpstreamKrewIndexRepo
	r.LocalKrewIndexRepoOwner = r.UpstreamKrewIndexRepoOwner
	r.LocalKrewIndexRepoCloneURL = r.UpstreamKrewIndexRepoCloneURL
}

func getUpstreamKrewIndexRepoCloneURL(gitProvider GitProvider, owner, repo string) (string, error) {
	override := os.Getenv("INPUT_UPSTREAM_KREW_INDEX_REPO_CLONE_URL")
	return gitProvider.ResolveCloneURL(owner, repo, override)
}
