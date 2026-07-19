package actions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
)

func getInput(key string) string {
	return os.Getenv("INPUT_" + strings.ToUpper(key))
}

func getTemplateFile(workdir string) string {
	templateFile := getInput(InputKeyKrewTemplateFile)
	if templateFile != "" {
		return filepath.Join(workdir, templateFile)
	}

	return filepath.Join(workdir, ".krew.yaml")
}

func shouldSubmitPRLocally() bool {
	return getInput(InputKeySubmitPRLocally) == "true"
}

func shouldDryRun() bool {
	return getInput(InputKeyDryRun) == "true"
}

func getIndexRepoConfig() releaser.IndexRepoConfig {
	upstreamRepoURL := getInput(InputKeyUpstreamKrewIndexRepoURL)
	if upstreamRepoURL == "" {
		upstreamRepoURL = "https://github.com/kubernetes-sigs/krew-index.git"
	}

	localRepoURL := getInput(InputKeyLocalKrewIndexRepoURL)
	if localRepoURL == "" {
		localRepoURL = "https://github.com/krew-release-bot/krew-index.git"
	}

	return releaser.IndexRepoConfigFromRaw(releaser.RawIndexRepoConfig{
		Upstream: releaser.RawReleaseTarget{
			ForgeKind: releaser.ForgeKindGitHub,
			RepoURL:   upstreamRepoURL,
			Auth: releaser.AuthConfig{
				Token:       os.Getenv("GITHUB_TOKEN"),
				TokenEnvVar: "GITHUB_TOKEN",
			},
		},
		LocalPushTarget: releaser.RawReleaseTarget{
			ForgeKind: releaser.ForgeKindGitHub,
			RepoURL:   localRepoURL,
			Auth: releaser.AuthConfig{
				Token:       os.Getenv("GITHUB_TOKEN"),
				TokenEnvVar: "GITHUB_TOKEN",
			},
		},
		DryRun: shouldDryRun(),
	})
}
