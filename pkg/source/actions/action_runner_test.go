package actions

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

type fakeReleaseRunner struct {
	releaseFunc func(request *source.ReleaseRequest) (string, error)
}

func (f *fakeReleaseRunner) Release(request *source.ReleaseRequest) (string, error) {
	return f.releaseFunc(request)
}

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
	restoreRetryOptions := source.SetRetryOptionsForTest(4, time.Millisecond, time.Millisecond)
	defer restoreRetryOptions()

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

func TestRunActionSubmitPRLocallyHappyPath(t *testing.T) {
	gock.OffAll()
	os.Clearenv()

	workdir := t.TempDir()
	templatePath := filepath.Join(workdir, "plugin.yaml")
	require.NoError(t, os.WriteFile(templatePath, []byte(validPluginManifest), 0644))

	os.Setenv("CIRCLECI", "true")
	os.Setenv("CIRCLE_TAG", "v1.2.3")
	os.Setenv("CIRCLE_PROJECT_USERNAME", "foo-bar")
	os.Setenv("CIRCLE_PROJECT_REPONAME", "my-awesome-plugin")
	os.Setenv("CIRCLE_USERNAME", "release-user")
	os.Setenv("INPUT_WORKDIR", workdir)
	os.Setenv("INPUT_KREW_TEMPLATE_FILE", "plugin.yaml")
	os.Setenv("INPUT_SUBMIT_PR_LOCALLY", "true")
	os.Setenv("INPUT_UPSTREAM_KREW_INDEX_REPO_URL", "git@github.example:org/platform/custom-index.git")
	os.Setenv("INPUT_LOCAL_KREW_INDEX_REPO_URL", "https://gitlab.example/mirror-group/team/mirror-index.git")
	os.Setenv("GITHUB_TOKEN", "token-123")

	originalFactory := newReleaseRunnerFromConfig
	defer func() { newReleaseRunnerFromConfig = originalFactory }()

	var capturedConfig releaser.IndexRepoConfig
	var capturedRequest *source.ReleaseRequest
	newReleaseRunnerFromConfig = func(config releaser.IndexRepoConfig) (releaseRunner, error) {
		capturedConfig = config
		return &fakeReleaseRunner{
			releaseFunc: func(request *source.ReleaseRequest) (string, error) {
				capturedRequest = request
				return "https://example.test/pr/123", nil
			},
		}, nil
	}

	err := RunAction()
	require.NoError(t, err)

	require.NotNil(t, capturedRequest)
	assert.Equal(t, "v1.2.3", capturedRequest.TagName)
	assert.Equal(t, "foo-bar", capturedRequest.PluginOwner)
	assert.Equal(t, "my-awesome-plugin", capturedRequest.PluginRepo)
	assert.Equal(t, "release-user", capturedRequest.PluginReleaseActor)
	assert.Equal(t, "whoami", capturedRequest.PluginName)
	assert.NotEmpty(t, capturedRequest.ProcessedTemplate)

	assert.Equal(t, releaser.ForgeKindGitHub, capturedConfig.Upstream.ForgeKind)
	assert.Equal(t, "org/platform/custom-index", capturedConfig.Upstream.Repo.FullPath())
	assert.Equal(t, "git@github.example:org/platform/custom-index.git", capturedConfig.Upstream.GitCloneURL)
	assert.Equal(t, "mirror-group/team/mirror-index", capturedConfig.LocalPushTarget.Repo.FullPath())
	assert.Equal(t, "https://gitlab.example/mirror-group/team/mirror-index.git", capturedConfig.LocalPushTarget.GitCloneURL)
	assert.Equal(t, "token-123", capturedConfig.Upstream.Auth.Token)
	assert.Equal(t, "token-123", capturedConfig.LocalPushTarget.Auth.Token)
	assert.False(t, capturedConfig.DryRun)
}

func TestRunActionSubmitPRLocallyDryRun(t *testing.T) {
	gock.OffAll()
	os.Clearenv()

	workdir := t.TempDir()
	templatePath := filepath.Join(workdir, "plugin.yaml")
	require.NoError(t, os.WriteFile(templatePath, []byte(validPluginManifest), 0644))

	os.Setenv("CIRCLECI", "true")
	os.Setenv("CIRCLE_TAG", "v1.2.3")
	os.Setenv("CIRCLE_PROJECT_USERNAME", "foo-bar")
	os.Setenv("CIRCLE_PROJECT_REPONAME", "my-awesome-plugin")
	os.Setenv("CIRCLE_USERNAME", "release-user")
	os.Setenv("INPUT_WORKDIR", workdir)
	os.Setenv("INPUT_KREW_TEMPLATE_FILE", "plugin.yaml")
	os.Setenv("INPUT_SUBMIT_PR_LOCALLY", "true")
	os.Setenv("INPUT_DRY_RUN", "true")
	os.Setenv("GITHUB_TOKEN", "token-123")

	originalFactory := newReleaseRunnerFromConfig
	defer func() { newReleaseRunnerFromConfig = originalFactory }()

	var capturedConfig releaser.IndexRepoConfig
	var capturedRequest *source.ReleaseRequest
	newReleaseRunnerFromConfig = func(config releaser.IndexRepoConfig) (releaseRunner, error) {
		capturedConfig = config
		return &fakeReleaseRunner{
			releaseFunc: func(request *source.ReleaseRequest) (string, error) {
				capturedRequest = request
				return "dry-run", nil
			},
		}, nil
	}

	err := RunAction()
	require.NoError(t, err)

	require.NotNil(t, capturedRequest)
	assert.Equal(t, "whoami", capturedRequest.PluginName)
	assert.True(t, capturedConfig.DryRun)
}

func TestRunActionWebhookDryRun(t *testing.T) {
	restoreRetryOptions := source.SetRetryOptionsForTest(4, time.Millisecond, time.Millisecond)
	defer restoreRetryOptions()

	gock.DisableNetworking()
	defer gock.OffAll()

	os.Clearenv()
	setupEnvironment()
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

	err := RunAction()
	require.NoError(t, err)
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

const validPluginManifest = `apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: whoami
spec:
  version: v0.0.6
  homepage: https://github.com/rajatjindal/kubectl-whoami
  platforms:
  - selector:
      matchLabels:
        os: darwin
        arch: amd64
    uri: https://github.com/rajatjindal/kubectl-whoami/releases/download/v0.0.6/darwin-amd64-v0.0.6.tar.gz
    sha256: f31e2237fdfd18467d8b5a391cb31f9fab70e9ef104e8618916025daa50489d5
    files:
    - from: "*"
      to: "."
    bin: kubectl-whoami
  shortDescription: Show the subject that's currently authenticated as.
  description: |
    This plugin show the subject that's currently authenticated as.
`
