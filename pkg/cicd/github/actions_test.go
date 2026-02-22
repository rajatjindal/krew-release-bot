package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name: "GITHUB_REF is in incorrect format",
			setup: func() {
				os.Setenv("GITHUB_REF", "tags/v5.0.0")
			},
			expectedError: `GITHUB_REF expected to be of format refs/tags/<tag> but found "tags/v5.0.0"`,
		},
		{
			name:          "GITHUB_REF is not found in env",
			expectedError: `GITHUB_REF env variable not found`,
		},
		{
			name: "krew_plugin_release_tag is provided",
			setup: func() {
				os.Setenv("INPUT_KREW_PLUGIN_RELEASE_TAG", "refs/tags/v5.0.0")
			},
			expectedTag: "v5.0.0",
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

func setupEnvironment() {
	os.Setenv("GITHUB_REPOSITORY", "foo-bar/my-awesome-plugin")
	os.Setenv("GITHUB_ACTOR", "karthik-aryan")
	os.Setenv("GITHUB_REF", "refs/tags/v0.0.2")
	os.Setenv("GITHUB_WORKSPACE", "./data/")
}
