package krew

import "os"

const (
	krewIndexRepoName  = "krew-index"
	krewIndexRepoOwner = "kubernetes-sigs"
)

// CloneURLResolver resolves the clone URL for an owner/repo pair.
type CloneURLResolver interface {
	ResolveCloneURL(owner, repo, override string) (string, error)
}

// GetKrewIndexRepoName returns the krew-index repo name
func GetKrewIndexRepoName() string {
	override := os.Getenv("INPUT_INDEX_REPO_NAME")
	if override != "" {
		return override
	}

	override = os.Getenv("UPSTREAM_KREW_INDEX_REPO_NAME")
	if override != "" {
		return override
	}

	return krewIndexRepoName
}

// GetKrewIndexRepoOwner returns the krew-index repo owner
func GetKrewIndexRepoOwner() string {
	override := os.Getenv("INPUT_INDEX_REPO_OWNER")
	if override != "" {
		return override
	}

	override = os.Getenv("UPSTREAM_KREW_INDEX_REPO_OWNER")
	if override != "" {
		return override
	}

	return krewIndexRepoOwner
}

// GetKrewIndexRepoCloneURL returns the clone URL for the target index repo.
func GetKrewIndexRepoCloneURL(resolver CloneURLResolver, owner, repo string) (string, error) {
	override := os.Getenv("INPUT_INDEX_REPO_CLONE_URL")
	return resolver.ResolveCloneURL(owner, repo, override)
}
