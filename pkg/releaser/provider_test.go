package releaser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
)

func TestGitHubProviderResolveCloneURL(t *testing.T) {
	provider := &gitHubGitProvider{}

	testcases := []struct {
		name      string
		override  string
		expected  string
		wantError string
	}{
		{
			name:     "uses default clone url",
			expected: "https://github.com/kubernetes-sigs/krew-index.git",
		},
		{
			name:     "accepts matching github clone url",
			override: "https://github.com/kubernetes-sigs/krew-index.git",
			expected: "https://github.com/kubernetes-sigs/krew-index.git",
		},
		{
			name:      "rejects non github host",
			override:  "https://evil.example/kubernetes-sigs/krew-index.git",
			wantError: "github clone url host must be github.com",
		},
		{
			name:      "rejects mismatched path",
			override:  "https://github.com/other-org/other-repo.git",
			wantError: "github clone url path must match /kubernetes-sigs/krew-index.git",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := provider.ResolveCloneURL("kubernetes-sigs", "krew-index", tc.override)
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("ResolveCloneURL returned error: %v", err)
			}

			if actual != tc.expected {
				t.Fatalf("unexpected clone url: %s", actual)
			}
		})
	}
}

func TestProviderRegistrationIsCaseInsensitive(t *testing.T) {
	RegisterGitProvider("StashGit", func() GitProvider {
		return &gitHubGitProvider{}
	})
	RegisterPRProvider("StashPR", func() PRProvider {
		return &gitHubPRProvider{}
	})

	gitProvider, err := getGitProvider("stashgit")
	if err != nil {
		t.Fatalf("getGitProvider returned error: %v", err)
	}
	if gitProvider.Name() != ProviderGitHub {
		t.Fatalf("unexpected git provider: %s", gitProvider.Name())
	}

	prProvider, err := getPRProvider("stashpr")
	if err != nil {
		t.Fatalf("getPRProvider returned error: %v", err)
	}
	if prProvider.Name() != ProviderGitHub {
		t.Fatalf("unexpected pr provider: %s", prProvider.Name())
	}
}

type testGitProvider struct {
	override string
	err      error
}

func (p testGitProvider) Name() string { return "test" }
func (p testGitProvider) DefaultCloneURL(owner, repo string) string {
	return "test://" + owner + "/" + repo
}
func (p testGitProvider) ResolveCloneURL(owner, repo, override string) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	if override != "" {
		return override, nil
	}
	return p.DefaultCloneURL(owner, repo), nil
}
func (p testGitProvider) GetAuth(_, _ string) transport.AuthMethod { return nil }

func TestGetUpstreamKrewIndexRepoCloneURL(t *testing.T) {
	testcases := []struct {
		name     string
		setup    func()
		expected string
		wantErr  string
		provider GitProvider
	}{
		{
			name:     "defaults to provider clone url",
			expected: "test://kubernetes-sigs/krew-index",
			provider: testGitProvider{},
		},
		{
			name: "new input override is set",
			setup: func() {
				os.Setenv("INPUT_UPSTREAM_KREW_INDEX_REPO_CLONE_URL", "ssh://example/custom-index.git")
			},
			expected: "ssh://example/custom-index.git",
			provider: testGitProvider{},
		},
		{
			name:     "provider errors are returned",
			wantErr:  assert.AnError.Error(),
			provider: testGitProvider{err: assert.AnError},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			if tc.setup != nil {
				tc.setup()
			}

			actual, err := getUpstreamKrewIndexRepoCloneURL(tc.provider, "kubernetes-sigs", "krew-index")
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
