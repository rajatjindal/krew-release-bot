//go:build integration

package actions

import (
	"os"
	"path/filepath"
	"testing"
)

type liveDryRunScenario struct {
	Name         string
	RepoSlug     string
	Actor        string
	Tag          string
	WorkDir      string
	TemplateFile string
}

func TestRunActionDryRunLive(t *testing.T) {
	if os.Getenv("KREW_RELEASE_BOT_LIVE") != "1" {
		t.Skip("set KREW_RELEASE_BOT_LIVE=1 to run live dry-run scenarios")
	}

	scenarios := []liveDryRunScenario{
		{
			Name:         "env-configured",
			RepoSlug:     requiredEnv(t, "KREW_RELEASE_BOT_LIVE_REPO"),
			Actor:        envOrDefault("KREW_RELEASE_BOT_LIVE_ACTOR", "local-user"),
			Tag:          envOrDefault("KREW_RELEASE_BOT_LIVE_TAG", "v0.0.1"),
			WorkDir:      requiredAbsPath(t, "KREW_RELEASE_BOT_LIVE_WORKDIR"),
			TemplateFile: envOrDefault("KREW_RELEASE_BOT_LIVE_TEMPLATE_FILE", ".krew.yaml"),
		},
		// Add more scenarios here when you want fixed live checks in code.
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Setenv("GITHUB_ACTIONS", "true")
			t.Setenv("GITHUB_REPOSITORY", scenario.RepoSlug)
			t.Setenv("GITHUB_ACTOR", scenario.Actor)
			t.Setenv("GITHUB_REF", "refs/tags/"+scenario.Tag)
			t.Setenv("GITHUB_WORKSPACE", scenario.WorkDir)
			t.Setenv("INPUT_KREW_TEMPLATE_FILE", scenario.TemplateFile)
			t.Setenv("INPUT_DRY_RUN", "true")

			if token := os.Getenv("KREW_RELEASE_BOT_LIVE_GITHUB_TOKEN"); token != "" {
				t.Setenv("GITHUB_TOKEN", token)
			}

			if err := RunAction(); err != nil {
				t.Fatalf("RunAction() error = %v", err)
			}
		})
	}
}

func requiredEnv(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("required env %s is not set", key)
	}

	return value
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func requiredAbsPath(t *testing.T, key string) string {
	t.Helper()

	value := requiredEnv(t, key)
	abs, err := filepath.Abs(value)
	if err != nil {
		t.Fatalf("filepath.Abs(%q): %v", value, err)
	}

	return abs
}
