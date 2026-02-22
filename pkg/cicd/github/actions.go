package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Actions implements provider interface
type Actions struct{}

func getHTTPClient() *http.Client {
	if os.Getenv("GITHUB_TOKEN") != "" {
		logrus.Info("GITHUB_TOKEN env variable found, using authenticated requests.")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
		return oauth2.NewClient(context.TODO(), ts)
	}

	return nil
}

func (p *Actions) getTagForCommitSha(commit string) (string, error) {
	client := github.NewClient(getHTTPClient())
	owner, repo, err := p.GetOwnerAndRepo()
	if err != nil {
		return "", err
	}

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, nil)
	if err != nil {
		return "", err
	}

	for _, release := range releases {
		if release.GetTargetCommitish() == commit {
			return release.GetTagName(), nil
		}
	}
	return "", fmt.Errorf("failed to find a release on this specific commit %q", commit)
}

// GetTag returns tag
func (p *Actions) GetTag() (string, error) {
	// check if user provided the tag, if not fallback to finding it from GITHUB_REF
	ref := getInputForAction("krew_plugin_release_tag")
	if ref == "" {
		ref = os.Getenv("GITHUB_REF")
		if ref == "" {
			return "", fmt.Errorf("GITHUB_REF env variable not found")
		}
	}

	//GITHUB_REF=refs/tags/v0.0.6
	if !strings.HasPrefix(ref, "refs/tags/") {
		return strings.ReplaceAll(ref, "refs/tags/", ""), nil
	}

	if strings.HasPrefix(ref, "refs/heads/") {
		return p.getTagForCommitSha(os.Getenv("GITHUB_SHA"))
	}

	return "", fmt.Errorf("failed to find the tag for the release")
}

// GetOwnerAndRepo gets the owner and repo from the env
func (p *Actions) GetOwnerAndRepo() (string, string, error) {
	repoFromEnv := os.Getenv("GITHUB_REPOSITORY")
	if repoFromEnv == "" {
		return "", "", fmt.Errorf("env GITHUB_REPOSITORY not set")
	}

	s := strings.Split(repoFromEnv, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("env GITHUB_REPOSITORY is incorrect format. expected format <owner>/<repo>, found %q", repoFromEnv)
	}

	return s[0], s[1], nil
}

// GetActor gets the owner and repo from the env
func (p *Actions) GetActor() (string, error) {
	actor := os.Getenv("GITHUB_ACTOR")
	if actor == "" {
		return "", fmt.Errorf("env GITHUB_ACTOR not set")
	}

	return actor, nil
}

// getInputForAction gets input to action
func getInputForAction(key string) string {
	return os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(key)))
}

// GetWorkDirectory gets workdir
func (p *Actions) GetWorkDirectory() string {
	workdirInput := getInputForAction("workdir")
	if workdirInput != "" {
		return workdirInput
	}

	return os.Getenv("GITHUB_WORKSPACE")
}

// GetTemplateFile returns the template file
func (p *Actions) GetTemplateFile() string {
	templateFile := getInputForAction("krew_template_file")
	if templateFile != "" {
		return filepath.Join(p.GetWorkDirectory(), templateFile)
	}

	return filepath.Join(p.GetWorkDirectory(), ".krew.yaml")
}
