package releaser

import (
	"testing"

	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeForge struct {
	currentUserFunc       func() (UserIdentity, error)
	repoDefaultBranchFunc func(repo RepoIdentity) (string, error)
	openPullRequestFunc   func(repo RepoIdentity, spec PullRequestSpec) (string, error)
}

func (f *fakeForge) CurrentUser() (UserIdentity, error) {
	return f.currentUserFunc()
}

func (f *fakeForge) RepoDefaultBranch(repo RepoIdentity) (string, error) {
	return f.repoDefaultBranchFunc(repo)
}

func (f *fakeForge) OpenPullRequest(repo RepoIdentity, spec PullRequestSpec) (string, error) {
	return f.openPullRequestFunc(repo, spec)
}

func TestNewReleaserFromConfig(t *testing.T) {
	forge := &fakeForge{
		currentUserFunc: func() (UserIdentity, error) {
			return UserIdentity{
				Handle: "bot-user",
				Name:   "Bot User",
				Email:  "bot@example.com",
			}, nil
		},
	}

	config := IndexRepoConfig{
		Upstream: ReleaseTarget{
			Repo: NewRepoIdentity("github.com", "kubernetes-sigs", "krew-index"),
			Auth: AuthConfig{Token: "upstream-token"},
		},
		LocalPushTarget: ReleaseTarget{
			Repo:      RepoIdentity{Host: "gitlab.example", Path: "group/subgroup/forked-krew-index", Owner: "mirror-user", Name: "forked-krew-index"},
			Auth:      AuthConfig{Token: "local-token"},
			ForgeKind: "gitlab",
		},
	}

	releaser, err := newReleaserWithForge(forge, config)
	require.NoError(t, err)

	assert.Equal(t, "local-token", releaser.Token)
	assert.Equal(t, "bot-user", releaser.TokenUserHandle)
	assert.Equal(t, "Bot User", releaser.TokenUsername)
	assert.Equal(t, "bot@example.com", releaser.TokenEmail)
	assert.Equal(t, config, releaser.Config)
	assert.Same(t, forge, releaser.forge)
}

func TestSubmitPRUsesUpstreamConfigAndLocalHeadOwner(t *testing.T) {
	request := &source.ReleaseRequest{
		TagName:            "v1.2.3",
		PluginName:         "kubectl-example",
		PluginOwner:        "rajatjindal",
		PluginRepo:         "kubectl-example",
		PluginReleaseActor: "release-user",
	}

	forge := &fakeForge{
		repoDefaultBranchFunc: func(repo RepoIdentity) (string, error) {
			assert.Equal(t, "kubernetes-sigs", repo.RepoOwner())
			assert.Equal(t, "krew-index", repo.RepoName())
			return "main", nil
		},
		openPullRequestFunc: func(repo RepoIdentity, spec PullRequestSpec) (string, error) {
			assert.Equal(t, "kubernetes-sigs", repo.RepoOwner())
			assert.Equal(t, "krew-index", repo.RepoName())
			assert.Equal(t, "release new version v1.2.3 of kubectl-example", spec.Title)
			assert.Equal(t, "rajatjindal-kubectl-example-kubectl-example-v1.2.3", spec.HeadBranch)
			assert.Equal(t, RepoIdentity{Host: "gitlab.example", Path: "mirror-user/forked-krew-index", Owner: "mirror-user", Name: "forked-krew-index"}, spec.HeadRepo)
			assert.Equal(t, "main", spec.BaseBranch)
			assert.Contains(t, spec.Body, "version `v1.2.3`")
			assert.Contains(t, spec.Body, "of `kubectl-example`")
			assert.Contains(t, spec.Body, "on behalf of @release-user")
			return "https://example.test/pr/1", nil
		},
	}

	releaser := &Releaser{
		Config: IndexRepoConfig{
			Upstream: ReleaseTarget{
				Repo: NewRepoIdentity("github.com", "kubernetes-sigs", "krew-index"),
			},
			LocalPushTarget: ReleaseTarget{
				Repo: RepoIdentity{Host: "gitlab.example", Path: "mirror-user/forked-krew-index", Owner: "mirror-user", Name: "forked-krew-index"},
			},
		},
		forge: forge,
	}

	prURL, err := releaser.submitPR(request)
	require.NoError(t, err)
	assert.Equal(t, "https://example.test/pr/1", prURL)
}

func TestSubmitPRUsesBaseBranchOverride(t *testing.T) {
	request := &source.ReleaseRequest{
		TagName:     "v1.2.3",
		PluginName:  "kubectl-example",
		PluginOwner: "rajatjindal",
		PluginRepo:  "kubectl-example",
	}

	forge := &fakeForge{
		repoDefaultBranchFunc: func(repo RepoIdentity) (string, error) {
			t.Fatalf("RepoDefaultBranch should not be called when BaseBranchOverride is set")
			return "", nil
		},
		openPullRequestFunc: func(repo RepoIdentity, spec PullRequestSpec) (string, error) {
			assert.Equal(t, "release-branch", spec.BaseBranch)
			return "https://example.test/pr/2", nil
		},
	}

	releaser := &Releaser{
		Config: IndexRepoConfig{
			Upstream: ReleaseTarget{
				Repo: NewRepoIdentity("github.com", "kubernetes-sigs", "krew-index"),
			},
			LocalPushTarget: ReleaseTarget{
				Repo: RepoIdentity{Host: "gitlab.example", Path: "mirror-user/forked-krew-index", Owner: "mirror-user", Name: "forked-krew-index"},
			},
			BaseBranchOverride: "release-branch",
		},
		forge: forge,
	}

	_, err := releaser.submitPR(request)
	require.NoError(t, err)
}

func TestRepoIdentityDerivesOwnerAndNameFromPath(t *testing.T) {
	repo := RepoIdentity{
		Host: "gitlab.example",
		Path: "group/subgroup/project-name",
	}

	assert.Equal(t, "group", repo.RepoOwner())
	assert.Equal(t, "project-name", repo.RepoName())
	assert.Equal(t, "group/subgroup/project-name", repo.FullPath())
}

func TestParseRepoIdentity(t *testing.T) {
	testcases := []struct {
		name         string
		repoURL      string
		expectedHost string
		expectedPath string
	}{
		{
			name:         "https github url",
			repoURL:      "https://github.com/kubernetes-sigs/krew-index.git",
			expectedHost: "github.com",
			expectedPath: "kubernetes-sigs/krew-index",
		},
		{
			name:         "ssh gitlab url",
			repoURL:      "git@gitlab.example:group/subgroup/project.git",
			expectedHost: "gitlab.example",
			expectedPath: "group/subgroup/project",
		},
		{
			name:         "bitbucket ssh url",
			repoURL:      "ssh://git@bitbucket.example/workspace/project.git",
			expectedHost: "bitbucket.example",
			expectedPath: "workspace/project",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := ParseRepoIdentity(ForgeKindGitHub, tc.repoURL)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedHost, repo.Host)
			assert.Equal(t, tc.expectedPath, repo.FullPath())
			assert.Equal(t, lastPathSegment(tc.expectedPath), repo.RepoName())
		})
	}
}
