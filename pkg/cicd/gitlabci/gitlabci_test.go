package gitlabci

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
			name: "CI_PROJECT_NAMESPACE and CI_PROJECT_NAME is set as expected",
			setup: func() {
				os.Setenv("CI_PROJECT_NAMESPACE", "foo-bar")
				os.Setenv("CI_PROJECT_NAME", "my-awesome-repo")
			},
			expectedOwner: "foo-bar",
			expectedRepo:  "my-awesome-repo",
		},
		{
			name: "CI_PROJECT_NAMESPACE is not set",
			setup: func() {
				os.Unsetenv("CI_PROJECT_NAMESPACE")
				os.Setenv("CI_PROJECT_NAME", "my-awesome-repo")
			},
			expectedError: `env CI_PROJECT_NAMESPACE not set`,
		},
		{
			name: "CI_PROJECT_NAME is not set",
			setup: func() {
				os.Setenv("CI_PROJECT_NAMESPACE", "foo-bar")
				os.Unsetenv("CI_PROJECT_NAME")
			},
			expectedError: `env CI_PROJECT_NAME not set`,
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
			name: "env GITLAB_USER_LOGIN is set as expected",
			setup: func() {
				os.Setenv("GITLAB_USER_LOGIN", "foo-bar")
			},
			expectedActor: "foo-bar",
		},
		{
			name:          "env GITLAB_USER_LOGIN is not set",
			expectedError: "env GITLAB_USER_LOGIN not set",
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
			name: "env CI_COMMIT_TAG is setup",
			setup: func() {
				os.Setenv("CI_COMMIT_TAG", "v5.0.0")
			},
			expectedTag: "v5.0.0",
		},
		{
			name: "CI_COMMIT_TAG is not set",
			setup: func() {
				os.Unsetenv("CI_COMMIT_TAG")
			},
			expectedError: `CI_COMMIT_TAG env variable not found`,
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

func TestGetWorkingDirectory(t *testing.T) {
	testcases := []struct {
		name        string
		setup       func()
		expectedDir string
	}{
		{
			name: "env CI_PROJECT_DIR is setup",
			setup: func() {
				os.Setenv("CI_PROJECT_DIR", "/builds/foo-bar/my-awesome-repo")
			},
			expectedDir: "/builds/foo-bar/my-awesome-repo",
		}, {
			name: "env CI_PROJECT_DIR is setup",
			setup: func() {
				os.Setenv("CI_PROJECT_DIR", "/builds/foo-bar/my-awesome-repo")
				os.Setenv("INPUT_WORKDIR", "/my-workdir")
			},
			expectedDir: "/my-workdir",
		}, {
			name: "env CI_PROJECT_DIR is not set",
			setup: func() {
				os.Unsetenv("CI_PROJECT_DIR")
			},
			expectedDir: "",
		},
	}

	p := &Provider{}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()

			if tc.setup != nil {
				tc.setup()
			}

			dir := p.GetWorkDirectory()
			assert.Equal(t, tc.expectedDir, dir)
		})
	}
}

func TestGetTemplateFile(t *testing.T) {
	testcases := []struct {
		name       string
		setup      func()
		expectFile string
	}{
		{
			name: "env INPUT_KREW_TEMPLATE_FILE is setup",
			setup: func() {
				os.Setenv("INPUT_KREW_TEMPLATE_FILE", "my-awesome-plugin.yaml")
			},
			expectFile: "my-awesome-plugin.yaml",
		}, {
			name: "env INPUT_KREW_TEMPLATE_FILE is not set",
			setup: func() {
				os.Unsetenv("INPUT_KREW_TEMPLATE_FILE")
			},
			expectFile: ".krew.yaml",
		},
	}

	p := &Provider{}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()

			if tc.setup != nil {
				tc.setup()
			}

			dir := p.GetTemplateFile()
			assert.Equal(t, tc.expectFile, dir)
		})
	}
}
