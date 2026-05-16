package krew

import "os"

const (
	KrewIndexRepoName = "krew-index"
	// not exporting intentionally
	krewIndexRepoOwner = "kubernetes-sigs"
)

// GetUpstreamKrewIndexRepoName returns the upstream krew-index repo name
func GetUpstreamKrewIndexRepoName() string {
	override := os.Getenv("UPSTREAM_KREW_INDEX_REPO_NAME")
	if override != "" {
		return override
	}

	return KrewIndexRepoName
}

// GetUpstreamKrewIndexRepoOwner returns the krew-index repo owner
func GetUpstreamKrewIndexRepoOwner() string {
	override := os.Getenv("UPSTREAM_KREW_INDEX_REPO_OWNER")
	if override != "" {
		return override
	}

	return krewIndexRepoOwner
}
