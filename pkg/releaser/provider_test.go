package releaser

import "testing"

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
