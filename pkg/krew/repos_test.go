package krew

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetKrewIndexRepoName(t *testing.T) {
	testcases := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name:     "env variable is not set",
			expected: "krew-index",
		},
		{
			name: "env variable is set to empty value",
			setup: func() {
				os.Setenv("UPSTREAM_KREW_INDEX_REPO_NAME", "")
			},
			expected: "krew-index",
		},
		{
			name: "new action input is set to value",
			setup: func() {
				os.Setenv("INPUT_UPSTREAM_KREW_INDEX_REPO_NAME", "new-custom-index")
			},
			expected: "new-custom-index",
		},
		{
			name: "env variable is set to value",
			setup: func() {
				os.Setenv("UPSTREAM_KREW_INDEX_REPO_NAME", "foo-bar")
			},
			expected: "foo-bar",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			if tc.setup != nil {
				tc.setup()
			}

			actual := GetKrewIndexRepoName()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetKrewIndexRepoOwner(t *testing.T) {
	testcases := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name:     "env variable is not set",
			expected: "kubernetes-sigs",
		},
		{
			name: "env variable is set to empty value",
			setup: func() {
				os.Setenv("UPSTREAM_KREW_INDEX_REPO_OWNER", "")
			},
			expected: "kubernetes-sigs",
		},
		{
			name: "new action input is set to value",
			setup: func() {
				os.Setenv("INPUT_UPSTREAM_KREW_INDEX_REPO_OWNER", "new-custom-owner")
			},
			expected: "new-custom-owner",
		},
		{
			name: "env variable is set to value",
			setup: func() {
				os.Setenv("UPSTREAM_KREW_INDEX_REPO_OWNER", "foo-bar-user")
			},
			expected: "foo-bar-user",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			if tc.setup != nil {
				tc.setup()
			}

			actual := GetKrewIndexRepoOwner()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetKrewIndexRepoCloneURL(t *testing.T) {
	testcases := []struct {
		name     string
		setup    func()
		expected string
		wantErr  string
		resolver CloneURLResolver
	}{
		{
			name:     "defaults to provider clone url",
			expected: "test://kubernetes-sigs/krew-index",
			resolver: testCloneURLResolver{},
		},
		{
			name: "new input override is set",
			setup: func() {
				os.Setenv("INPUT_UPSTREAM_KREW_INDEX_REPO_CLONE_URL", "ssh://example/custom-index.git")
			},
			expected: "ssh://example/custom-index.git",
			resolver: testCloneURLResolver{},
		},
		{
			name:     "resolver errors are returned",
			wantErr:  assert.AnError.Error(),
			resolver: failingCloneURLResolver{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			if tc.setup != nil {
				tc.setup()
			}

			actual, err := GetKrewIndexRepoCloneURL(tc.resolver, GetKrewIndexRepoOwner(), GetKrewIndexRepoName())
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

type testCloneURLResolver struct{}

func (r testCloneURLResolver) ResolveCloneURL(owner, repo, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	return "test://" + owner + "/" + repo, nil
}

type failingCloneURLResolver struct{}

func (r failingCloneURLResolver) ResolveCloneURL(_, _, _ string) (string, error) {
	return "", assert.AnError
}
