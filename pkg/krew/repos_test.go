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
				os.Setenv("upstream-krew-index-repo-name", "")
			},
			expected: "krew-index",
		},
		{
			name: "env variable is set to value",
			setup: func() {
				os.Setenv("upstream-krew-index-repo-name", "foo-bar")
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
			expected: "rajatjin",
		},
		{
			name: "env variable is set to empty value",
			setup: func() {
				os.Setenv("upstream-krew-index-repo-owner", "")
			},
			expected: "rajatjin",
		},
		{
			name: "env variable is set to value",
			setup: func() {
				os.Setenv("upstream-krew-index-repo-owner", "foo-bar-user")
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
