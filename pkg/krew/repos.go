package krew

import "os"

const (
	krewIndexRepoName          = "krew-index"
	krewIndexRepoOwner         = "kubernetes-sigs"
	upstreamKrewIndexRepoName  = "UPSTREAM_KREW_INDEX_REPO_NAME"
	upstreamKrewIndexRepoOwner = "UPSTREAM_KREW_INDEX_REPO_OWNER"
)

// GetKrewIndexRepoName returns the krew-index repo name
func GetKrewIndexRepoName() string {
	override := os.Getenv(upstreamKrewIndexRepoName)
	if override != "" {
		return override
	}

	return krewIndexRepoName
}

// GetKrewIndexRepoOwner returns the krew-index repo owner
func GetKrewIndexRepoOwner() string {
	override := os.Getenv(upstreamKrewIndexRepoOwner)
	if override != "" {
		return override
	}

	return krewIndexRepoOwner
}

func SetKrewIndexRepoName(name string) string {
	if name != "" && os.Getenv(upstreamKrewIndexRepoName) == "" {
		return name
	}
	return GetKrewIndexRepoName()
}

func SetKrewIndexRepoOwner(owner string) string {
	if owner != "" && os.Getenv(upstreamKrewIndexRepoOwner) == "" {
		return owner
	}
	return GetKrewIndexRepoName()
}
