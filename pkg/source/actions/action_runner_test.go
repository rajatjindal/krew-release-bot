package actions

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func assertError(t *testing.T, expectedError string, err error) {
	if expectedError == "" {
		assert.Nil(t, err)
	}

	if expectedError != "" {
		assert.NotNil(t, err)
		if err != nil {
			assert.Equal(t, expectedError, err.Error())
		}
	}
}

func TestRunAction(t *testing.T) {
	restore := source.SetRetryWaitForTests(func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	})
	defer restore()

	testcases := []struct {
		name          string
		setup         func()
		setupMocks    func()
		expectedError string
	}{
		{
			name: "no release tag found",
			setup: func() {
				os.Setenv("GITHUB_REF", "")
			},
			expectedError: "GITHUB_REF env variable not found",
		},
		{
			name: "no release info found for the tag",
			setup: func() {
				gock.New("https://api.github.com").
					Times(4).
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(404).
					BodyString("no release with tag v0.0.2 found")
			},
			expectedError: "GET https://api.github.com/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2: 404  []",
		},
		{
			name: "owner and repo not found",
			setup: func() {
				os.Setenv("GITHUB_REPOSITORY", "")
			},
			expectedError: `env GITHUB_REPOSITORY not set`,
		},
		{
			name: "actor not found",
			setup: func() {
				os.Setenv("GITHUB_ACTOR", "")
			},
			expectedError: `env GITHUB_ACTOR not set`,
		},
		{
			name: "release is a pre-release",
			setup: func() {
				gock.New("https://api.github.com").
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(200).
					BodyString(preRelease)
			},
			expectedError: `release with tag "v0.0.2" is a pre-release. skipping`,
		},
		{
			name: "release have assets, but downloading them fails",
			setup: func() {
				gock.New("https://api.github.com").
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(200).
					BodyString(releaseWithAssets)

				gock.New("https://github.com").
					Times(4).
					Get("/foo-bar/my-awesome-plugin/releases/download/v0.0.2/darwin-amd64-v0.0.2.tar.gz").
					Reply(404).
					BodyString("darwin-amd64-v0.0.2.tar.gz not found")

				gock.New("https://github.com").
					Get("/foo-bar/my-awesome-plugin/releases/download/v0.0.2/linux-amd64-v0.0.2.tar.gz").
					Reply(200).
					BodyString("linux-amd64")
			},
			expectedError: `template: .krew.yaml:13:6: executing ".krew.yaml" at <addURIAndSha "https://github.com/foo-bar/my-awesome-plugin/releases/download/{{ .TagName }}/darwin-amd64-{{ .TagName }}.tar.gz" .TagName>: error calling addURIAndSha: downloading file https://github.com/foo-bar/my-awesome-plugin/releases/download/v0.0.2/darwin-amd64-v0.0.2.tar.gz failed. status code: 404, expected: 200`,
		},
		{
			name: "release have assets",
			setup: func() {
				gock.New("https://api.github.com").
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(200).
					BodyString(releaseWithAssets)

				gock.New("https://github.com").
					Get("/foo-bar/my-awesome-plugin/releases/download/v0.0.2/darwin-amd64-v0.0.2.tar.gz").
					Reply(200).
					BodyString("darwin-amd64-v0.0.2.tar.gz")

				gock.New("https://github.com").
					Get("/foo-bar/my-awesome-plugin/releases/download/v0.0.2/linux-amd64-v0.0.2.tar.gz").
					Reply(200).
					BodyString("linux-amd64")

				gock.New("https://krew-release-bot.rajatjindal.com").
					Post("/github-action-webhook").
					Reply(200).
					JSON("PR https://github.com/kubernetes-sigs/krew-index/pull/26 opened successfully")

			},
		},
		{
			name: "dry run skips webhook submission",
			setup: func() {
				os.Setenv("INPUT_DRY_RUN", "true")
				gock.New("https://api.github.com").
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(200).
					BodyString(releaseWithAssets)

				gock.New("https://github.com").
					Get("/foo-bar/my-awesome-plugin/releases/download/v0.0.2/darwin-amd64-v0.0.2.tar.gz").
					Reply(200).
					BodyString("darwin-amd64-v0.0.2.tar.gz")

				gock.New("https://github.com").
					Get("/foo-bar/my-awesome-plugin/releases/download/v0.0.2/linux-amd64-v0.0.2.tar.gz").
					Reply(200).
					BodyString("linux-amd64")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gock.DisableNetworking()

			//reset env
			os.Clearenv()
			setupEnvironment()

			if tc.setup != nil {
				tc.setup()
			}

			err := RunAction()
			assertError(t, tc.expectedError, err)
			logrus.Error(gock.GetUnmatchedRequests())

			for _, g := range gock.GetUnmatchedRequests() {
				logrus.Infof("UNMATCHED => %#v", g)
			}

			gock.OffAll()
		})
	}
}

type repoConfigSnapshot struct {
	owner    string
	repo     string
	cloneURL string
}

type directPRExpectation struct {
	gitProvider string
	prProvider  string
	token       string
	pluginName  string
	upstream    repoConfigSnapshot
	directPush  repoConfigSnapshot
}

func TestSubmitReleaseRequestUsesWebhookByDefault(t *testing.T) {
	gock.DisableNetworking()
	defer gock.OffAll()

	gock.New("https://example.com").
		Post("/webhook").
		Reply(200).
		JSON("ok")

	original := runDirectPRRelease
	defer func() {
		runDirectPRRelease = original
	}()
	runDirectPRRelease = func(config directPRConfig, request *source.ReleaseRequest) (string, error) {
		t.Fatalf("direct PR path should not be used for default webhook mode")
		return "", nil
	}

	pr, err := submitReleaseRequest(&source.ReleaseRequest{PluginName: "test-plugin"}, actionConfig{
		WebhookURL: "https://example.com/webhook",
	})
	assert.NoError(t, err)
	assert.Equal(t, "ok", pr)
}

func TestSubmitReleaseRequestUsesDirectPRWhenConfigured(t *testing.T) {
	original := runDirectPRRelease
	defer func() {
		runDirectPRRelease = original
	}()

	testcases := []struct {
		name    string
		config  actionConfig
		request *source.ReleaseRequest
		want    directPRExpectation
	}{
		{
			name: "custom repo override",
			config: actionConfig{
				DirectPR: directPRConfig{
					Token:       "top-secret-token",
					GitProvider: "github",
					PRProvider:  "github",
					TargetRepo: releaser.IndexRepoConfig{
						Owner:    "acme",
						Name:     "custom-index",
						CloneURL: "https://github.com/acme/custom-index.git",
					},
				},
			},
			request: &source.ReleaseRequest{PluginName: "test-plugin"},
			want: directPRExpectation{
				gitProvider: "github",
				prProvider:  "github",
				token:       "top-secret-token",
				pluginName:  "test-plugin",
				upstream: repoConfigSnapshot{
					owner:    "acme",
					repo:     "custom-index",
					cloneURL: "https://github.com/acme/custom-index.git",
				},
				directPush: repoConfigSnapshot{
					owner:    "acme",
					repo:     "custom-index",
					cloneURL: "https://github.com/acme/custom-index.git",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			runDirectPRRelease = func(config directPRConfig, request *source.ReleaseRequest) (string, error) {
				assertDirectPRExpectation(t, tc.want, config, request)
				return "ok", nil
			}

			pr, err := submitReleaseRequest(tc.request, tc.config)
			assert.NoError(t, err)
			assert.Equal(t, "ok", pr)
		})
	}
}

func TestSubmitReleaseRequestModeSelection(t *testing.T) {
	original := runDirectPRRelease
	defer func() {
		runDirectPRRelease = original
	}()

	testcases := []struct {
		name              string
		config            actionConfig
		wantWebhookCalls  int
		wantDirectPRCalls int
	}{
		{
			name:              "default behavior uses webhook",
			config:            actionConfig{WebhookURL: "https://example.com/webhook"},
			wantWebhookCalls:  1,
			wantDirectPRCalls: 0,
		},
		{
			name: "direct PR token switches behavior",
			config: actionConfig{
				WebhookURL: "https://example.com/webhook",
				DirectPR: directPRConfig{
					Token:       "top-secret-token",
					GitProvider: "github",
					PRProvider:  "github",
					TargetRepo: releaser.IndexRepoConfig{
						Owner: "acme",
						Name:  "custom-index",
					},
				},
			},
			wantWebhookCalls:  0,
			wantDirectPRCalls: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gock.DisableNetworking()
			defer gock.OffAll()

			webhookCalls := 0
			directPRCalls := 0

			gock.New("https://example.com").
				Post("/webhook").
				Reply(200).
				Map(func(response *http.Response) *http.Response {
					webhookCalls++
					return response
				}).
				JSON("ok")

			runDirectPRRelease = func(config directPRConfig, request *source.ReleaseRequest) (string, error) {
				directPRCalls++
				return "ok", nil
			}

			pr, err := submitReleaseRequest(&source.ReleaseRequest{PluginName: "test-plugin"}, tc.config)
			assert.NoError(t, err)
			assert.Equal(t, "ok", pr)
			assert.Equal(t, tc.wantWebhookCalls, webhookCalls)
			assert.Equal(t, tc.wantDirectPRCalls, directPRCalls)
		})
	}
}

func assertDirectPRExpectation(t *testing.T, want directPRExpectation, config directPRConfig, request *source.ReleaseRequest) {
	t.Helper()

	assert.Equal(t, want.gitProvider, config.GitProvider)
	assert.Equal(t, want.prProvider, config.PRProvider)
	assert.Equal(t, want.token, config.Token)
	assert.Equal(t, want.pluginName, request.PluginName)

	r, err := releaser.NewWithProviders(config.GitProvider, config.PRProvider, config.Token, config.TargetRepo)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assertRepoConfig(t, want.upstream, r.UpstreamKrewIndexRepoOwner, r.UpstreamKrewIndexRepo, r.UpstreamKrewIndexRepoCloneURL)

	r.ConfigureDirectPRs()

	assertRepoConfig(t, want.directPush, r.LocalKrewIndexRepoOwner, r.LocalKrewIndexRepo, r.LocalKrewIndexRepoCloneURL)
}

func assertRepoConfig(t *testing.T, want repoConfigSnapshot, owner, repo, cloneURL string) {
	t.Helper()
	assert.Equal(t, want.owner, owner)
	assert.Equal(t, want.repo, repo)
	assert.Equal(t, want.cloneURL, cloneURL)
}

const preRelease = `{
	"id": 22569944,
	"tag_name": "v0.0.2",
	"name": "v0.0.2",
	"prerelease": true
}`

const releaseWithAssets = `{
	"id": 22569944,
	"tag_name": "v0.0.2",
	"name": "v0.0.2",
	"prerelease": false,
	"assets": [
		{
			"id": 16605457,
			"node_id": "MDEyOlJlbGVhc2VBc3NldDE2NjA1NDU3",
			"name": "darwin-amd64-v0.0.2.tar.gz"
		},
		{
			"id": 16605458,
			"node_id": "MDEyOlJlbGVhc2VBc3NldDE2NjA1NDU3",
			"name": "linux-amd64-v0.0.2.tar.gz"
		}
	]
}`

func setupEnvironment() {
	os.Setenv("GITHUB_REPOSITORY", "foo-bar/my-awesome-plugin")
	os.Setenv("GITHUB_ACTOR", "karthik-aryan")
	os.Setenv("GITHUB_REF", "refs/tags/v0.0.2")
	os.Setenv("GITHUB_WORKSPACE", "./data/")
	os.Setenv("GITHUB_ACTIONS", "true")
}
