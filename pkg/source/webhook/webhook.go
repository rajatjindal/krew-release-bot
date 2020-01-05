package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
)

//GithubWebhook is github webhook handler
type GithubWebhook struct {
	webhookSecret string
}

//NewGithubWebhook gets new git webhook instance
func NewGithubWebhook(webhookSecret string) (*GithubWebhook, error) {
	return &GithubWebhook{
		webhookSecret: webhookSecret,
	}, nil
}

//Parse validates and parse the webhook request into a release request
func (gw *GithubWebhook) Parse(r *http.Request) (*source.ReleaseRequest, error) {
	body, err := gw.isValidSignature(r)
	if err != nil {
		return nil, err
	}

	t := github.WebHookType(r)
	if t != "" {
		return nil, fmt.Errorf("expected a release event, got %q", t)
	}

	e, err := github.ParseWebHook(t, body)
	if err != nil {
		return nil, fmt.Errorf("failed to parsepayload. error: %v", err)
	}

	event, ok := e.(*github.ReleaseEvent)
	if !ok {
		return nil, fmt.Errorf("expected a release event, got %T", event)
	}

	if event.GetAction() != "published" {
		return nil, fmt.Errorf("expected release event action 'published', got %q", event.GetAction())
	}

	if event.GetRelease().GetPrerelease() {
		return nil, fmt.Errorf("release with tag %q is a pre-release. skipping", event.GetRelease().GetTagName())
	}

	if len(event.GetRelease().Assets) == 0 {
		return nil, fmt.Errorf("no assets found for release with tag %q", event.GetRelease().GetTagName())
	}

	templateFile, err := getKrewTemplate(
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		event.GetRelease().GetTagName(),
	)

	if err != nil {
		return nil, err
	}

	releaseRequest := &source.ReleaseRequest{
		TagName:            event.GetRelease().GetTagName(),
		PluginOwner:        event.GetRepo().GetOwner().GetLogin(),
		PluginRepo:         event.GetRepo().GetName(),
		PluginReleaseActor: event.GetSender().GetLogin(),
		TemplateFile:       templateFile,
	}

	pluginName, pluginManifest, err := source.ProcessTemplate(templateFile, releaseRequest)
	if err != nil {
		return nil, err
	}

	releaseRequest.PluginName = pluginName
	releaseRequest.ProcessedTemplate = pluginManifest

	return releaseRequest, nil
}

//TODO: possibly use creds here for private repo scenario
func getKrewTemplate(owner, repo, tagName string) (string, error) {
	templateFileLoc := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/.krew.yaml", owner, repo, tagName)

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}

	file := filepath.Join(dir, ".krew.yaml")
	return source.DownloadFileWithName(templateFileLoc, file)
}

func (gw *GithubWebhook) isValidSignature(r *http.Request) ([]byte, error) {
	gotHash := strings.SplitN(r.Header.Get("X-Hub-Signature"), "=", 2)
	if gotHash[0] != "sha1" {
		return nil, fmt.Errorf("expected sha1 hash, got %q", gotHash[0])
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Cannot read the request body: %s", err)
	}

	hash := hmac.New(sha1.New, []byte(gw.webhookSecret))
	if _, err := hash.Write(b); err != nil {
		return nil, fmt.Errorf("Cannot compute the HMAC for request: %s", err)
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	if gotHash[1] != expectedHash {
		return nil, fmt.Errorf("expected hash %q, got %q", expectedHash, gotHash[1])
	}

	return b, nil
}
