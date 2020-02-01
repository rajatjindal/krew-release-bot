package travisci

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
			name: "TRAVIS_REPO_SLUG is set as expected",
			setup: func() {
				os.Setenv("TRAVIS_REPO_SLUG", "foo-bar/my-awesome-repo")
			},
			expectedOwner: "foo-bar",
			expectedRepo:  "my-awesome-repo",
		},
		{
			name: "TRAVIS_REPO_SLUG is set in incorrect format",
			setup: func() {
				os.Setenv("TRAVIS_REPO_SLUG", "foo-bar:my-awesome-repo")
			},
			expectedError: `env TRAVIS_REPO_SLUG is incorrect format. expected format <owner>/<repo>, found "foo-bar:my-awesome-repo"`,
		},
		{
			name: "TRAVIS_REPO_SLUG is set in incorrect format 2",
			setup: func() {
				os.Setenv("TRAVIS_REPO_SLUG", "my-awesome-repo")
			},
			expectedError: `env TRAVIS_REPO_SLUG is incorrect format. expected format <owner>/<repo>, found "my-awesome-repo"`,
		},
		{
			name:          "TRAVIS_REPO_SLUG environment is not set",
			expectedError: `env TRAVIS_REPO_SLUG not set`,
		},
	}

	p := &Provider{}
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

func TestGetActionActor(t *testing.T) {
	testcases := []struct {
		name          string
		setup         func()
		expectedActor string
		expectedError string
	}{
		{
			name: "env TRAVIS_REPO_SLUG is set as expected",
			setup: func() {
				os.Setenv("TRAVIS_REPO_SLUG", "foo-bar/my-awesome-plugin")
			},
			expectedActor: "foo-bar",
		},
		{
			name:          "env TRAVIS_REPO_SLUG is not set",
			expectedError: "env TRAVIS_REPO_SLUG not set",
		},
	}

	p := &Provider{}
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
			name: "env TRAVIS_TAG is setup",
			setup: func() {
				os.Setenv("TRAVIS_TAG", "v5.0.0")
			},
			expectedTag: "v5.0.0",
		},
		{
			name: "TRAVIS_TAG is not set",
			setup: func() {
				os.Unsetenv("TRAVIS_TAG")
			},
			expectedError: `TRAVIS_TAG env variable not found`,
		},
	}

	p := &Provider{}
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
	os.Setenv("TRAVIS_REPO_SLUG", "foo-bar/my-awesome-plugin")
	os.Setenv("TRAVIS_TAG", "v0.0.2")
	os.Setenv("TRAVIS_BUILD_DIR", "./data/")
}
