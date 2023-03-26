package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/rajatjindal/krew-release-bot/pkg/cicd"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func getHTTPClient() *http.Client {
	if os.Getenv("GITHUB_TOKEN") != "" {
		logrus.Info("GITHUB_TOKEN env variable found, using authenticated requests.")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
		return oauth2.NewClient(context.TODO(), ts)
	}

	return nil
}

// RunAction runs the github action
func RunAction() error {
	client := github.NewClient(getHTTPClient())
	provider := cicd.GetProvider()

	if provider == nil {
		logrus.Fatal("failed to identify the CI/CD provider")
	}

	tag, err := provider.GetTag()
	if err != nil {
		return err
	}

	owner, repo, err := provider.GetOwnerAndRepo()
	if err != nil {
		return err
	}

	actor, err := provider.GetActor()
	if err != nil {
		return err
	}

	releaseInfo, err := getReleaseForTag(client, owner, repo, tag)
	if err != nil {
		return err
	}

	if releaseInfo.GetPrerelease() {
		return fmt.Errorf("release with tag %q is a pre-release. skipping", releaseInfo.GetTagName())
	}

	templateFile := provider.GetTemplateFile()
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

func getReleaseForTag(client *github.Client, owner, repo, tag string) (*github.RepositoryRelease, error) {
	release, _, err := client.Repositories.GetReleaseByTag(context.TODO(), owner, repo, tag)
	if err != nil {
		return nil, err
	}

	return release, nil
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
	respBody, err := io.ReadAll(resp.Body)
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
