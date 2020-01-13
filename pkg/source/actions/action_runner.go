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
	client := github.NewClient(nil)

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

	templateFile := getTemplateFile()
	logrus.Infof("using template file %q", templateFile)

	releaseRequest := &source.ReleaseRequest{
		TagName:            releaseInfo.GetTagName(),
		PluginOwner:        owner,
		PluginRepo:         repo,
		PluginReleaseActor: actor,
		TemplateFile:       templateFile,
	}

	pluginName, pluginManifest, err := source.ProcessTemplate(templateFile, releaseRequest)
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

	req, err := http.NewRequest(http.MethodPost, getWebhookURL(), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

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

func getWebhookURL() string {
	if os.Getenv("KREW_RELEASE_BOT_WEBHOOK_URL") != "" {
		return os.Getenv("KREW_RELEASE_BOT_WEBHOOK_URL")
	}

	return "https://krew-release-bot.rajatjindal.com/github-action-webhook"
}

//getInputForAction gets input to action
func getInputForAction(key string) string {
	return os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(key)))
}

func getWorkDirectory() string {
	workdirInput := getInputForAction("workdir")
	if workdirInput != "" {
		return workdirInput
	}

	return os.Getenv("GITHUB_WORKSPACE")
}

func getTemplateFile() string {
	templateFile := getInputForAction("krew_template_file")
	if templateFile != "" {
		return filepath.Join(getWorkDirectory(), templateFile)
	}

	return filepath.Join(getWorkDirectory(), ".krew.yaml")
}
