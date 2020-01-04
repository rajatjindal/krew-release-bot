package krew

import "os"

const (
	krewIndexRepoName  = "krew-index"
	krewIndexRepoOwner = "kubernetes-sigs"
)

//GetKrewIndexRepoName returns the krew-index repo name
func GetKrewIndexRepoName() string {
	override := os.Getenv("UPSTREAM_KREW_INDEX_REPO_NAME")
	if override != "" {
		return override
	}

	return krewIndexRepoName
}

//GetKrewIndexRepoOwner returns the krew-index repo owner
func GetKrewIndexRepoOwner() string {
	override := os.Getenv("UPSTREAM_KREW_INDEX_REPO_OWNER")
	if override != "" {
		return override
	}

	return krewIndexRepoOwner
}
