package main

import (
	"os"
	"testing"

	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/stretchr/testify/assert"
)

func TestWebhookConfigDefaults(t *testing.T) {
	os.Clearenv()

	config := getIndexRepoConfigFromEnv()

	assert.Equal(t, releaser.ForgeKindGitHub, config.Upstream.ForgeKind)
	assert.Equal(t, "github.com", config.Upstream.Repo.Host)
	assert.Equal(t, "kubernetes-sigs/krew-index", config.Upstream.Repo.FullPath())
	assert.Equal(t, "kubernetes-sigs", config.Upstream.Repo.RepoOwner())
	assert.Equal(t, "krew-index", config.Upstream.Repo.RepoName())
	assert.Equal(t, "https://github.com/kubernetes-sigs/krew-index.git", config.Upstream.GitCloneURL)
	assert.Equal(t, "GH_TOKEN", config.Upstream.Auth.TokenEnvVar)

	assert.Equal(t, releaser.ForgeKindGitHub, config.LocalPushTarget.ForgeKind)
	assert.Equal(t, "github.com", config.LocalPushTarget.Repo.Host)
	assert.Equal(t, "krew-release-bot/krew-index", config.LocalPushTarget.Repo.FullPath())
	assert.Equal(t, "krew-release-bot", config.LocalPushTarget.Repo.RepoOwner())
	assert.Equal(t, "krew-index", config.LocalPushTarget.Repo.RepoName())
	assert.Equal(t, "https://github.com/krew-release-bot/krew-index.git", config.LocalPushTarget.GitCloneURL)
	assert.Equal(t, "GH_TOKEN", config.LocalPushTarget.Auth.TokenEnvVar)
}

func TestWebhookConfigOverrides(t *testing.T) {
	os.Clearenv()
	os.Setenv("UPSTREAM_KREW_INDEX_FORGE_KIND", "github")
	os.Setenv("UPSTREAM_KREW_INDEX_API_BASE_URL", "https://github.example/api/v3/")
	os.Setenv("UPSTREAM_KREW_INDEX_REPO_URL", "git@github.example:org/platform/custom-index.git")
	os.Setenv("LOCAL_KREW_INDEX_FORGE_KIND", "gitlab")
	os.Setenv("LOCAL_KREW_INDEX_API_BASE_URL", "https://gitlab.example/api/v4/")
	os.Setenv("LOCAL_KREW_INDEX_REPO_URL", "https://gitlab.example/mirror-group/team/mirror-index.git")
	os.Setenv("LOCAL_KREW_INDEX_TOKEN", "local-token")
	os.Setenv("KREW_INDEX_BASE_BRANCH", "stable")

	config := getIndexRepoConfigFromEnv()

	assert.Equal(t, releaser.ForgeKindGitHub, config.Upstream.ForgeKind)
	assert.Equal(t, "https://github.example/api/v3/", config.Upstream.APIBaseURL)
	assert.Equal(t, "github.example", config.Upstream.Repo.Host)
	assert.Equal(t, "org/platform/custom-index", config.Upstream.Repo.FullPath())
	assert.Equal(t, "org", config.Upstream.Repo.RepoOwner())
	assert.Equal(t, "custom-index", config.Upstream.Repo.RepoName())
	assert.Equal(t, "git@github.example:org/platform/custom-index.git", config.Upstream.GitCloneURL)

	assert.Equal(t, releaser.ForgeKind("gitlab"), config.LocalPushTarget.ForgeKind)
	assert.Equal(t, "https://gitlab.example/api/v4/", config.LocalPushTarget.APIBaseURL)
	assert.Equal(t, "gitlab.example", config.LocalPushTarget.Repo.Host)
	assert.Equal(t, "mirror-group/team/mirror-index", config.LocalPushTarget.Repo.FullPath())
	assert.Equal(t, "mirror-group", config.LocalPushTarget.Repo.RepoOwner())
	assert.Equal(t, "mirror-index", config.LocalPushTarget.Repo.RepoName())
	assert.Equal(t, "https://gitlab.example/mirror-group/team/mirror-index.git", config.LocalPushTarget.GitCloneURL)
	assert.Equal(t, "local-token", config.LocalPushTarget.Auth.Token)
	assert.Equal(t, "LOCAL_KREW_INDEX_TOKEN", config.LocalPushTarget.Auth.TokenEnvVar)
	assert.Equal(t, "stable", config.BaseBranchOverride)
}
