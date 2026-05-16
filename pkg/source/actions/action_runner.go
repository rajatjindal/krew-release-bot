package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rajatjindal/krew-release-bot/pkg/cicd"
	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
)

type RunOptions struct {
	DryRun *bool
}

type releaseMetadata struct {
	Tag          string
	PluginOwner  string
	PluginRepo   string
	Actor        string
	TemplateFile string
}

type directPRConfig struct {
	Token       string
	GitProvider string
	PRProvider  string
	TargetRepo  releaser.IndexRepoConfig
}

type actionConfig struct {
	DryRun     bool
	WebhookURL string
	DirectPR   directPRConfig
}

var runDirectPRRelease = func(config directPRConfig, request *source.ReleaseRequest) (string, error) {
	r, err := releaser.NewWithProviders(config.GitProvider, config.PRProvider, config.Token, config.TargetRepo)
	if err != nil {
		return "", err
	}
	r.ConfigureDirectPRs()
	return r.Release(request)
}

// RunAction runs the github action
func RunAction() error {
	return RunActionWithOptions(RunOptions{})
}

func RunActionWithOptions(options RunOptions) error {
	provider, err := cicd.GetProvider()
	if err != nil {
		return err
	}

	if provider == nil {
		return fmt.Errorf("failed to identify the CI/CD provider")
	}

	metadata, err := resolveReleaseMetadata(provider)
	if err != nil {
		return err
	}

	config := resolveActionConfig(options)

	// this currently works only for GitHub.
	// for travisci and circleci it always return false, nil
	prerelease, err := provider.IsPreRelease(metadata.PluginOwner, metadata.PluginRepo, metadata.Tag)
	if err != nil {
		return err
	}

	if prerelease {
		return fmt.Errorf("release with tag %q is a pre-release. skipping", metadata.Tag)
	}

	templateFile := metadata.TemplateFile
	logrus.Infof("using template file %q", templateFile)

	releaseRequest := &source.ReleaseRequest{
		TagName:            metadata.Tag,
		PluginOwner:        metadata.PluginOwner,
		PluginRepo:         metadata.PluginRepo,
		PluginReleaseActor: metadata.Actor,
		TemplateFile:       templateFile,
	}

	pluginName, pluginManifest, err := source.ProcessTemplate(templateFile, releaseRequest)
	if err != nil {
		return err
	}

	releaseRequest.PluginName = pluginName
	releaseRequest.ProcessedTemplate = pluginManifest

	if config.DryRun {
		return logDryRun(releaseRequest, config)
	}

	pr, err := submitReleaseRequest(releaseRequest, config)
	if err != nil {
		return err
	}

	logrus.Info(pr)
	return nil
}

func resolveReleaseMetadata(provider cicd.Provider) (releaseMetadata, error) {
	tag, err := provider.GetTag()
	if err != nil {
		return releaseMetadata{}, err
	}

	owner, repo, err := provider.GetOwnerAndRepo()
	if err != nil {
		return releaseMetadata{}, err
	}

	actor, err := provider.GetActor()
	if err != nil {
		return releaseMetadata{}, err
	}

	return releaseMetadata{
		Tag:          tag,
		PluginOwner:  owner,
		PluginRepo:   repo,
		Actor:        actor,
		TemplateFile: provider.GetTemplateFile(),
	}, nil
}

func resolveActionConfig(options RunOptions) actionConfig {
	dryRun := isDryRunEnabled()
	if options.DryRun != nil {
		dryRun = *options.DryRun
	}

	directPR := directPRConfig{
		Token:       getInputForAction("upstream_krew_index_repo_token"),
		GitProvider: getInputForAction("upstream_krew_index_repo_provider"),
		PRProvider:  getInputForAction("upstream_krew_index_pr_provider"),
		TargetRepo: releaser.IndexRepoConfig{
			Owner:    firstNonEmpty(getInputForAction("upstream_krew_index_repo_owner"), krew.DefaultIndexRepoOwner),
			Name:     firstNonEmpty(getInputForAction("upstream_krew_index_repo_name"), krew.DefaultIndexRepoName),
			CloneURL: getInputForAction("upstream_krew_index_repo_clone_url"),
		},
	}

	if directPR.GitProvider == "" {
		directPR.GitProvider = releaser.ProviderGitHub
	}
	if directPR.PRProvider == "" {
		directPR.PRProvider = directPR.GitProvider
	}

	return actionConfig{
		DryRun:     dryRun,
		WebhookURL: getWebhookURL(),
		DirectPR:   directPR,
	}
}

func logDryRun(request *source.ReleaseRequest, config actionConfig) error {
	body, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return err
	}

	mode := "webhook"
	if config.DirectPR.Token != "" {
		mode = "direct-pr"
	}

	logrus.Infof("dry-run enabled, skipping %s submission", mode)
	fmt.Println(string(body))
	return nil
}

func submitReleaseRequest(request *source.ReleaseRequest, config actionConfig) (string, error) {
	if config.DirectPR.Token == "" {
		if hasCustomTargetRepo(config.DirectPR) {
			return "", fmt.Errorf("custom upstream krew index repo configuration requires upstream_krew_index_repo_token so the PR can be opened directly from CI")
		}

		return submitForPR(request, config.WebhookURL)
	}

	logrus.Infof(
		"upstream_krew_index_repo_token provided, opening PR directly from CI using git provider %q and pr provider %q",
		config.DirectPR.GitProvider,
		config.DirectPR.PRProvider,
	)
	return runDirectPRRelease(config.DirectPR, request)
}

func hasCustomTargetRepo(config directPRConfig) bool {
	owner := firstNonEmpty(config.TargetRepo.Owner, krew.DefaultIndexRepoOwner)
	name := firstNonEmpty(config.TargetRepo.Name, krew.DefaultIndexRepoName)
	gitProvider := firstNonEmpty(config.GitProvider, releaser.ProviderGitHub)
	prProvider := firstNonEmpty(config.PRProvider, gitProvider)

	return owner != krew.DefaultIndexRepoOwner ||
		name != krew.DefaultIndexRepoName ||
		config.TargetRepo.CloneURL != "" ||
		gitProvider != releaser.ProviderGitHub ||
		prProvider != releaser.ProviderGitHub
}

func submitForPR(request *source.ReleaseRequest, webhookURL string) (string, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
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

func isDryRunEnabled() bool {
	value := strings.TrimSpace(strings.ToLower(getInputForAction("dry_run")))
	return value == "1" || value == "true" || value == "yes"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
