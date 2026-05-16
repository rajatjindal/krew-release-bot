package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rajatjindal/krew-release-bot/pkg/cicd"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/rajatjindal/krew-release-bot/pkg/types"
	"github.com/sirupsen/logrus"
)

// RunAction runs the github action
func RunAction() error {
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

	releaseRequest := &types.ReleaseRequest{
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

	if true {
		pr, err := submitPRViaWebhook(releaseRequest)
		if err != nil {
			return err
		}
		logrus.Info(pr)
	} else {
		pr, err := submitPR(releaseRequest)
		if err != nil {
			return err
		}
		logrus.Info(pr)
	}

	return nil
}

func submitPRViaWebhook(request *types.ReleaseRequest) (string, error) {
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

func submitPR(request *types.ReleaseRequest) (string, error) {
	releaserInst, err := releaser.New("")
	if err != nil {
		return "", err
	}

	pr, err := releaserInst.Release(request)
	if err != nil {
		return "", errors.Wrap(err, "opening pr")
	}

	return fmt.Sprintf("PR %q submitted successfully", pr), nil
}
