package actions

import (
	"os"
	"testing"

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

const releaseNoAssets = `{
	"id": 22569944,
	"tag_name": "v0.0.2",
	"name": "v0.0.2",
	"prerelease": false,
	"assets": []
}`

func setupEnvironment() {
	os.Setenv("GITHUB_REPOSITORY", "foo-bar/my-awesome-plugin")
	os.Setenv("GITHUB_ACTOR", "karthik-aryan")
	os.Setenv("GITHUB_REF", "refs/tags/v0.0.2")
	os.Setenv("GITHUB_WORKSPACE", "./data/")
	os.Setenv("GITHUB_ACTIONS", "true")
}
