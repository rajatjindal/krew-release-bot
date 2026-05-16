package releaser

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v66/github"
	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"golang.org/x/oauth2"
)

func getUserDetails(token string) (string, string, string, error) {
	if token == "" {
		return "", "", "", fmt.Errorf("no token provided")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch authenticated github user details from token")
	}

	return user.GetLogin(), user.GetName(), user.GetEmail(), nil
}

// getLocalKrewIndexRepoOwner returns the local krew-index repo name
func getLocalKrewIndexRepoName() string {
	override := os.Getenv("LOCAL_KREW_INDEX_REPO_NAME")
	if override != "" {
		return override
	}

	return krew.KrewIndexRepoName
}

// getLocalKrewIndexRepoOwner returns the local krew-index repo owner
func getLocalKrewIndexRepo(tokenUserHandle, repoName string) string {
	override := os.Getenv("LOCAL_KREW_INDEX_REPO")
	if override != "" {
		return override
	}

	return getCloneURL(tokenUserHandle, repoName)
}

func getCloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}
