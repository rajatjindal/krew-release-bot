package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
)

//RunAction runs the github action
func RunAction() error {
	mc := &http.Client{
		Transport: &authInjector{token: os.Getenv("GITHUB_TOKEN")},
	}
	client := github.NewClient(mc)

	tag, err := getTag()
	if err != nil {
		return err
	}

	owner, repo, err := getOwnerAndRepo()
	if err != nil {
		return err
	}

	actor, err := getActionActor()
	if err != nil {
		return err
	}

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

	pr, err := submitForPR(releaseRequest)
	if err != nil {
		return err
	}

	logrus.Info(pr)
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
	owner, repo, err := getOwnerAndRepo()
	if err != nil {
		return nil, err
	}

	release, _, err := client.Repositories.GetReleaseByTag(context.TODO(), owner, repo, tag)
	if err != nil {
		return nil, err
	}

	return release, nil
}

//getOwnerAndRepo gets the owner and repo from the env
func getOwnerAndRepo() (string, string, error) {
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

//getActionActor gets the owner and repo from the env
func getActionActor() (string, error) {
	actor := os.Getenv("GITHUB_ACTOR")
	if actor == "" {
		return "", fmt.Errorf("env GITHUB_ACTOR not set")
	}

	return actor, nil
}

func submitForPR(request *source.ReleaseRequest) (string, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "https://krew-release-bot-test.rajatjindal.com/github-action-webhook", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Add("x-github-token", os.Getenv("GITHUB_TOKEN"))
	req.Header.Add("content-type", "application/json")

	client := http.Client{
		Timeout: time.Duration(30 * time.Second),
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected status code %d got %d. body: %s", http.StatusOK, resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}
