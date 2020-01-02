package krew

import "os"

const (
	krewIndexRepoName  = "krew-index"
	krewIndexRepoOwner = "rajatjin"
)

//GetKrewIndexRepoName returns the krew-index repo name
func GetKrewIndexRepoName() string {
	override := os.Getenv("upstream-krew-index-repo-name")
	if override != "" {
		return override
	}

	return krewIndexRepoName
}

//GetKrewIndexRepoOwner returns the krew-index repo owner
func GetKrewIndexRepoOwner() string {
	override := os.Getenv("upstream-krew-index-repo-owner")
	if override != "" {
		return override
	}

	return krewIndexRepoOwner
}
