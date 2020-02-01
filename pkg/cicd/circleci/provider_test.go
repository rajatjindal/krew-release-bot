package circleci

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
			name: "CIRCLE_PROJECT_USERNAME and CIRCLE_PROJECT_REPONAME is set as expected",
			setup: func() {
				os.Setenv("CIRCLE_PROJECT_USERNAME", "foo-bar")
				os.Setenv("CIRCLE_PROJECT_REPONAME", "my-awesome-repo")
			},
			expectedOwner: "foo-bar",
			expectedRepo:  "my-awesome-repo",
		},
		{
			name: "CIRCLE_PROJECT_USERNAME is not set",
			setup: func() {
				os.Unsetenv("CIRCLE_PROJECT_USERNAME")
				os.Setenv("CIRCLE_PROJECT_REPONAME", "my-awesome-repo")
			},
			expectedError: `env CIRCLE_PROJECT_USERNAME not set`,
		},
		{
			name: "CIRCLE_PROJECT_REPONAME is not set",
			setup: func() {
				os.Setenv("CIRCLE_PROJECT_USERNAME", "foo-bar")
				os.Unsetenv("CIRCLE_PROJECT_REPONAME")
			},
			expectedError: `env CIRCLE_PROJECT_REPONAME not set`,
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
			name: "env CIRCLE_USERNAME is set as expected",
			setup: func() {
				os.Setenv("CIRCLE_USERNAME", "foo-bar")
			},
			expectedActor: "foo-bar",
		},
		{
			name:          "env CIRCLE_USERNAME is not set",
			expectedError: "env CIRCLE_USERNAME not set",
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
			name: "env CIRCLE_TAG is setup",
			setup: func() {
				os.Setenv("CIRCLE_TAG", "v5.0.0")
			},
			expectedTag: "v5.0.0",
		},
		{
			name: "CIRCLE_TAG is not set",
			setup: func() {
				os.Unsetenv("CIRCLE_TAG")
			},
			expectedError: `CIRCLE_TAG env variable not found`,
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
	os.Setenv("CIRCLE_PROJECT_USERNAME", "foo-bar")
	os.Setenv("CIRCLE_PROJECT_REPONAME", "my-awesome-plugin")
	os.Setenv("CIRCLE_USERNAME", "karthik-aryan")
	os.Setenv("CIRCLE_TAG", "v0.0.2")
	os.Setenv("CIRCLE_WORKING_DIRECTORY", "./data/")
}
