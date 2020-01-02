package actions

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty"
	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
)

//RunAction runs the github action
func RunAction() error {
	tag, err := getTag()
	if err != nil {
		return err
	}

	mc := &http.Client{
		Transport: &authInjector{token: os.Getenv("GITHUB_TOKEN")},
	}
	client := github.NewClient(mc)

	releaseInfo, err := getReleaseForTag(client, tag)
	if err != nil {
		return err
	}

	if releaseInfo.GetPrerelease() {
		return fmt.Errorf("release with tag %q is a pre-release. skipping", releaseInfo.GetTagName())
	}

	if len(releaseInfo.Assets) == 0 {
		return fmt.Errorf("no assets found for release with tag %q", tag)
	}

	owner, repo := getOwnerAndRepo()
	actor := getActionActor()

	releaseRequest := &source.ReleaseRequest{
		TagName:            releaseInfo.GetTagName(),
		PluginOwner:        owner,
		PluginRepo:         repo,
		PluginReleaseActor: actor,
		TemplateFile:       filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".krew.yaml"),
	}

	pluginName, pluginManifest, err := source.ProcessTemplate(filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".krew.yaml"), releaseRequest)
	if err != nil {
		return err
	}

	releaseRequest.PluginName = pluginName
	releaseRequest.ProcessedTemplate = pluginManifest

	err = submitForPR(releaseRequest)
	if err != nil {
		return err
	}

	return nil
}

func getTag() (string, error) {
	ref := os.Getenv("GITHUB_REF")
	if ref == "" {
		return "", fmt.Errorf("GITHUB_REF env variable not found")
	}

	//GITHUB_REF=refs/tags/v0.0.6
	if !strings.HasPrefix(ref, "refs/tags/") {
		return "", fmt.Errorf("GITHUB_REF expected to be of format refs/tags/<tag> but found %q", ref)
	}

	return strings.ReplaceAll(ref, "refs/tags/", ""), nil
}

func getReleaseForTag(client *github.Client, tag string) (*github.RepositoryRelease, error) {
	owner, repo := getOwnerAndRepo()
	release, _, err := client.Repositories.GetReleaseByTag(context.TODO(), owner, repo, tag)
	if err != nil {
		return nil, err
	}

	return release, nil
}

//getOwnerAndRepo gets the owner and repo from the env
func getOwnerAndRepo() (string, string) {
	s := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	return s[0], s[1]
}

//getActionActor gets the owner and repo from the env
func getActionActor() string {
	return os.Getenv("GITHUB_ACTOR")
}

func submitForPR(request *source.ReleaseRequest) error {
	client := resty.New()
	resp, err := client.R().
		SetBody(request).
		SetHeader("x-github-token", os.Getenv("GITHUB_TOKEN")).
		Post("https://krew-release-bot-test.rajatjindal.com/github-action-webhook")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("expected status code %d got %d. body: %s", http.StatusOK, resp.StatusCode(), resp.Body())
	}

	logrus.Info(string(resp.Body()))
	return nil
}
