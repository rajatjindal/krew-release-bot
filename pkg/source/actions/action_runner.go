package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rajatjindal/krew-release-bot/pkg/cicd"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
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
	provider := cicd.GetProvider()

	if provider == nil {
		return fmt.Errorf("failed to identify the CI/CD provider")
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

	// this currently works only for GitHub.
	// for travisci and circleci it always return false, nil
	prerelease, err := provider.IsPreRelease(owner, repo, tag)
	if err != nil {
		return err
	}

	if prerelease {
		return fmt.Errorf("release with tag %q is a pre-release. skipping", tag)
	}

	templateFile := provider.GetTemplateFile()
	logrus.Infof("using template file %q", templateFile)

	releaseRequest := &source.ReleaseRequest{
		TagName:            tag,
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

	pr, err := submitReleaseRequest(releaseRequest)
	if err != nil {
		return err
	}

	logrus.Info(pr)
	return nil
}

func submitReleaseRequest(request *source.ReleaseRequest) (string, error) {
	token := getInputForAction("upstream_krew_index_repo_token")
	if token == "" {
		if getInputForAction("upstream_krew_index_repo_owner") != "" ||
			getInputForAction("upstream_krew_index_repo_name") != "" ||
			getInputForAction("upstream_krew_index_repo_provider") != "" ||
			getInputForAction("upstream_krew_index_repo_clone_url") != "" {
			return "", fmt.Errorf("custom upstream krew index repo configuration requires upstream_krew_index_repo_token so the PR can be opened directly from CI")
		}

		return submitForPR(request)
	}

	providerName := getInputForAction("upstream_krew_index_repo_provider")
	if providerName == "" {
		providerName = releaser.ProviderGitHub
	}
	prProviderName := getInputForAction("upstream_krew_index_pr_provider")
	if prProviderName == "" {
		prProviderName = providerName
	}

	logrus.Infof(
		"upstream_krew_index_repo_token provided, opening PR directly from CI using git provider %q and pr provider %q",
		providerName,
		prProviderName,
	)
	r, err := releaser.NewWithProviders(providerName, prProviderName, token)
	if err != nil {
		return "", err
	}
	r.ConfigureDirectPRs()
	return r.Release(request)
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

func getInputForAction(key string) string {
	return os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(key)))
}
