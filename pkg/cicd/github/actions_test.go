package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestGetOwnerAndRepo(t *testing.T) {
	testcases := []struct {
		name          string
		setup         func()
		expectedOwner string
		expectedRepo  string
		expectedError string
	}{
		{
			name: "GITHUB_REPOSITORY is set as expected",
			setup: func() {
				os.Setenv("GITHUB_REPOSITORY", "foo-bar/my-awesome-repo")
			},
			expectedOwner: "foo-bar",
			expectedRepo:  "my-awesome-repo",
		},
		{
			name: "GITHUB_REPOSITORY is set in incorrect format",
			setup: func() {
				os.Setenv("GITHUB_REPOSITORY", "foo-bar:my-awesome-repo")
			},
			expectedError: `env GITHUB_REPOSITORY is incorrect format. expected format <owner>/<repo>, found "foo-bar:my-awesome-repo"`,
		},
		{
			name: "GITHUB_REPOSITORY is set in incorrect format 2",
			setup: func() {
				os.Setenv("GITHUB_REPOSITORY", "my-awesome-repo")
			},
			expectedError: `env GITHUB_REPOSITORY is incorrect format. expected format <owner>/<repo>, found "my-awesome-repo"`,
		},
		{
			name:          "GITHUB_REPOSITORY environment is not set",
			expectedError: `env GITHUB_REPOSITORY not set`,
		},
	}

	p := &Actions{}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()

			if tc.setup != nil {
				tc.setup()
			}

			owner, repo, err := p.GetOwnerAndRepo()

			assert.Equal(t, tc.expectedOwner, owner)
			assert.Equal(t, tc.expectedRepo, repo)
			assertError(t, tc.expectedError, err)
		})
	}
}

func TestGetActor(t *testing.T) {
	testcases := []struct {
		name          string
		setup         func()
		expectedActor string
		expectedError string
	}{
		{
			name: "env GITHUB_ACTOR is set as expected",
			setup: func() {
				os.Setenv("GITHUB_ACTOR", "foo-bar")
			},
			expectedActor: "foo-bar",
		},
		{
			name:          "env GITHUB_ACTOR is not set",
			expectedError: "env GITHUB_ACTOR not set",
		},
	}

	p := &Actions{}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()

			if tc.setup != nil {
				tc.setup()
			}

			actor, err := p.GetActor()
			assert.Equal(t, tc.expectedActor, actor)
			assertError(t, tc.expectedError, err)
		})
	}
}
func TestGetTag(t *testing.T) {
	testcases := []struct {
		name          string
		setup         func()
		expectedTag   string
		expectedError string
	}{
		{
			name: "env GITHUB_REF is setup as expected",
			setup: func() {
				os.Setenv("GITHUB_REF", "refs/tags/v5.0.0")
			},
			expectedTag: "v5.0.0",
		},
		{
			name: "env GITHUB_REF is setup as expected",
			setup: func() {
				os.Setenv("GITHUB_REF", "refs/heads/main")
				os.Setenv("GITHUB_REPOSITORY", "rajatjindal/kubectl-whoami")
				os.Setenv("GITHUB_SHA", "5de971cf57f0aecdc08018e712ca0845f2dac714")

				gock.New("https://api.github.com").
					Get("/repos/rajatjindal/kubectl-whoami/releases").
					Reply(200).
					BodyString(`[
  {
    "url": "https://api.github.com/repos/rajatjindal/kubectl-whoami/releases/1",
    "html_url": "https://github.com/rajatjindal/kubectl-whoami/releases/v1.0.0",
    "assets_url": "https://api.github.com/repos/rajatjindal/kubectl-whoami/releases/1/assets",
    "upload_url": "https://uploads.github.com/repos/rajatjindal/kubectl-whoami/releases/1/assets{?name,label}",
    "tarball_url": "https://api.github.com/repos/rajatjindal/kubectl-whoami/tarball/v1.0.0",
    "zipball_url": "https://api.github.com/repos/rajatjindal/kubectl-whoami/zipball/v1.0.0",
    "id": 1,
    "node_id": "MDc6UmVsZWFzZTE=",
    "tag_name": "v1.0.0",
    "target_commitish": "5de971cf57f0aecdc08018e712ca0845f2dac714",
    "name": "v1.0.0",
    "body": "Description of the release",
    "draft": false,
    "prerelease": false,
    "created_at": "2013-02-27T19:35:32Z",
    "published_at": "2013-02-27T19:35:32Z",
    "author": {
      "login": "rajatjindal",
      "id": 1,
      "node_id": "MDQ6VXNlcjE=",
      "avatar_url": "https://github.com/images/error/rajatjindal_happy.gif",
      "gravatar_id": "",
      "url": "https://api.github.com/users/rajatjindal",
      "html_url": "https://github.com/rajatjindal",
      "followers_url": "https://api.github.com/users/rajatjindal/followers",
      "following_url": "https://api.github.com/users/rajatjindal/following{/other_user}",
      "gists_url": "https://api.github.com/users/rajatjindal/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/rajatjindal/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/rajatjindal/subscriptions",
      "organizations_url": "https://api.github.com/users/rajatjindal/orgs",
      "repos_url": "https://api.github.com/users/rajatjindal/repos",
      "events_url": "https://api.github.com/users/rajatjindal/events{/privacy}",
      "received_events_url": "https://api.github.com/users/rajatjindal/received_events",
      "type": "User",
      "site_admin": false
    },
    "assets": [
      {
        "url": "https://api.github.com/repos/rajatjindal/kubectl-whoami/releases/assets/1",
        "browser_download_url": "https://github.com/rajatjindal/kubectl-whoami/releases/download/v1.0.0/example.zip",
        "id": 1,
        "node_id": "MDEyOlJlbGVhc2VBc3NldDE=",
        "name": "example.zip",
        "label": "short description",
        "state": "uploaded",
        "content_type": "application/zip",
        "size": 1024,
        "download_count": 42,
        "created_at": "2013-02-27T19:35:32Z",
        "updated_at": "2013-02-27T19:35:32Z",
        "uploader": {
          "login": "rajatjindal",
          "id": 1,
          "node_id": "MDQ6VXNlcjE=",
          "avatar_url": "https://github.com/images/error/rajatjindal_happy.gif",
          "gravatar_id": "",
          "url": "https://api.github.com/users/rajatjindal",
          "html_url": "https://github.com/rajatjindal",
          "followers_url": "https://api.github.com/users/rajatjindal/followers",
          "following_url": "https://api.github.com/users/rajatjindal/following{/other_user}",
          "gists_url": "https://api.github.com/users/rajatjindal/gists{/gist_id}",
          "starred_url": "https://api.github.com/users/rajatjindal/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/rajatjindal/subscriptions",
          "organizations_url": "https://api.github.com/users/rajatjindal/orgs",
          "repos_url": "https://api.github.com/users/rajatjindal/repos",
          "events_url": "https://api.github.com/users/rajatjindal/events{/privacy}",
          "received_events_url": "https://api.github.com/users/rajatjindal/received_events",
          "type": "User",
          "site_admin": false
        }
      }
    ]
  }
]`)
			},
			expectedTag: "v1.0.0",
		},
		{
			name: "GITHUB_REF is in incorrect format",
			setup: func() {
				os.Setenv("GITHUB_REF", "tags/v5.0.0")
			},
			expectedError: `failed to find the tag for the release`,
		},
		{
			name:          "GITHUB_REF is not found in env",
			expectedError: `GITHUB_REF env variable not found`,
		},
	}

	p := &Actions{}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()

			if tc.setup != nil {
				tc.setup()
			}

			tag, err := p.GetTag()
			assert.Equal(t, tc.expectedTag, tag)
			assertError(t, tc.expectedError, err)
		})
	}
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
