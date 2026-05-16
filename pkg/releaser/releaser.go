package releaser

import (
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

type IndexRepoConfig struct {
	Owner    string
	Name     string
	CloneURL string
}

// TODO: get email, userhandle, name from token
func getUserDetails(_ string) (string, string, string) {
	return "krew-release-bot", "Krew Release Bot", "krewpluginreleasebot@gmail.com"
}

// New returns new releaser object
func New(providerName, token string, indexRepo IndexRepoConfig) (*Releaser, error) {
	return NewWithProviders(providerName, providerName, token, indexRepo)
}

// NewWithProviders returns a releaser configured with separate git and PR providers.
func NewWithProviders(gitProviderName, prProviderName, token string, indexRepo IndexRepoConfig) (*Releaser, error) {
	gitProvider, err := getGitProvider(gitProviderName)
	if err != nil {
		return nil, err
	}
	prProvider, err := getPRProvider(prProviderName)
	if err != nil {
		return nil, err
	}

	tokenUserHandle, tokenUsername, tokenEmail := getUserDetails(token)

	indexRepo = withDefaultIndexRepoConfig(indexRepo)

	upstreamCloneURL, err := resolveUpstreamKrewIndexRepoCloneURL(gitProvider, indexRepo)
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
		UpstreamKrewIndexRepo:         indexRepo.Name,
		UpstreamKrewIndexRepoOwner:    indexRepo.Owner,
		UpstreamKrewIndexRepoCloneURL: upstreamCloneURL,
		LocalKrewIndexRepo:            indexRepo.Name,
		LocalKrewIndexRepoOwner:       tokenUserHandle,
		LocalKrewIndexRepoCloneURL:    gitProvider.DefaultCloneURL(tokenUserHandle, indexRepo.Name),
	}, nil
}

// ConfigureDirectPRs updates the releaser to push branches directly to the target index repo.
func (r *Releaser) ConfigureDirectPRs() {
	r.LocalKrewIndexRepo = r.UpstreamKrewIndexRepo
	r.LocalKrewIndexRepoOwner = r.UpstreamKrewIndexRepoOwner
	r.LocalKrewIndexRepoCloneURL = r.UpstreamKrewIndexRepoCloneURL
}

func withDefaultIndexRepoConfig(indexRepo IndexRepoConfig) IndexRepoConfig {
	if indexRepo.Owner == "" {
		indexRepo.Owner = krew.DefaultIndexRepoOwner
	}
	if indexRepo.Name == "" {
		indexRepo.Name = krew.DefaultIndexRepoName
	}

	return indexRepo
}

func resolveUpstreamKrewIndexRepoCloneURL(gitProvider GitProvider, indexRepo IndexRepoConfig) (string, error) {
	indexRepo = withDefaultIndexRepoConfig(indexRepo)
	return gitProvider.ResolveCloneURL(indexRepo.Owner, indexRepo.Name, indexRepo.CloneURL)
}
