package main

import (
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
)

const (
	krewIndexRepoName     = "krew-index"
	krewIndexRepoOwner    = "kubernetes-sigs"
	defaultLocalRepoOwner = "krew-release-bot"
)

func getIndexRepoConfigFromEnv() releaser.IndexRepoConfig {
	upstreamRepoURL := getEnvOrDefault("UPSTREAM_KREW_INDEX_REPO_URL", defaultRepoURL("github.com", krewIndexRepoOwner, krewIndexRepoName))
	localRepoURL := getEnvOrDefault("LOCAL_KREW_INDEX_REPO_URL", defaultRepoURL("github.com", defaultLocalRepoOwner, krewIndexRepoName))

	return releaser.IndexRepoConfigFromRaw(releaser.RawIndexRepoConfig{
		Upstream: releaser.RawReleaseTarget{
			ForgeKind:  releaser.ForgeKind(getEnvOrDefault("UPSTREAM_KREW_INDEX_FORGE_KIND", string(releaser.ForgeKindGitHub))),
			APIBaseURL: os.Getenv("UPSTREAM_KREW_INDEX_API_BASE_URL"),
			RepoURL:    upstreamRepoURL,
			Auth: releaser.AuthConfig{
				Token:       os.Getenv("GH_TOKEN"),
				TokenEnvVar: "GH_TOKEN",
			},
		},
		LocalPushTarget: releaser.RawReleaseTarget{
			ForgeKind:  releaser.ForgeKind(getEnvOrDefault("LOCAL_KREW_INDEX_FORGE_KIND", string(releaser.ForgeKindGitHub))),
			APIBaseURL: os.Getenv("LOCAL_KREW_INDEX_API_BASE_URL"),
			RepoURL:    localRepoURL,
			Auth: releaser.AuthConfig{
				Token:       getEnvOrDefault("LOCAL_KREW_INDEX_TOKEN", os.Getenv("GH_TOKEN")),
				TokenEnvVar: configuredEnvOrDefault("LOCAL_KREW_INDEX_TOKEN", "GH_TOKEN"),
			},
		},
		BaseBranchOverride: os.Getenv("KREW_INDEX_BASE_BRANCH"),
	})
}

func getEnvOrDefault(key, fallback string) string {
	if override := os.Getenv(key); override != "" {
		return override
	}

	return fallback
}

func configuredEnvOrDefault(primary, fallback string) string {
	if os.Getenv(primary) != "" {
		return primary
	}

	return fallback
}

func defaultRepoURL(host, owner, repo string) string {
	return "https://" + host + "/" + owner + "/" + repo + ".git"
}
