package github

import (
	"context"
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
		owner         string
		repo          string
		tag           string
		setup         func()
		expectedError string
	}{
		{
			name:  "release exists",
			owner: "foo-bar",
			repo:  "my-awesome-plugin",
			tag:   "v0.0.2",
			setup: func() {
				gock.New("https://api.github.com").
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(200).
					BodyString(releaseWithAssets)
			},
		},
		{
			name:  "release is a pre-release",
			owner: "foo-bar",
			repo:  "my-awesome-plugin",
			tag:   "v0.0.2",
			setup: func() {
				gock.New("https://api.github.com").
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(200).
					BodyString(preRelease)
			},
			expectedError: `release with tag "v0.0.2" is a pre-release. skipping`,
		},
		{
			name:  "release does not exist",
			owner: "foo-bar",
			repo:  "my-awesome-plugin",
			tag:   "v0.0.2",
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
			name:  "invalid token",
			owner: "foo-bar",
			repo:  "my-awesome-plugin",
			tag:   "v0.0.2",
			setup: func() {
				os.Setenv("GITHUB_TOKEN", "12345")
				gock.New("https://api.github.com").
					Times(4).
					MatchHeader("Authorization", "^Bearer 12345$").
					Get("/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2").
					Reply(401).
					BodyString("no release with tag v0.0.2 found")
			},
			expectedError: "GET https://api.github.com/repos/foo-bar/my-awesome-plugin/releases/tags/v0.0.2: 401  []",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gock.DisableNetworking()

			if tc.setup != nil {
				tc.setup()
			}

			validator := &Validator{}
			err := validator.Validate(context.Background(), tc.owner, tc.repo, tc.tag)
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
