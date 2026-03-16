package gitlabci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Provider implements provider interface
type Provider struct{}

func (p *Provider) IsPreRelease(owner, repo, tag string) (bool, error) {
	// gitlab doesn't have Prerelease.
	return false, nil
}

// GetTag returns tag
func (p *Provider) GetTag() (string, error) {
	ref := os.Getenv("CI_COMMIT_TAG")
	if ref == "" {
		return "", fmt.Errorf("CI_COMMIT_TAG env variable not found")
	}

	return ref, nil
}

// GetOwnerAndRepo gets the owner and repo from the env
func (p *Provider) GetOwnerAndRepo() (string, string, error) {
	owner := os.Getenv("CI_PROJECT_NAMESPACE")
	if owner == "" {
		return "", "", fmt.Errorf("env CI_PROJECT_NAMESPACE not set")
	}

	repo := os.Getenv("CI_PROJECT_NAME")
	if repo == "" {
		return "", "", fmt.Errorf("env CI_PROJECT_NAME not set")
	}

	return owner, repo, nil
}

// GetActor gets the owner and repo from the env
func (p *Provider) GetActor() (string, error) {
	actor := os.Getenv("GITLAB_USER_LOGIN")
	if actor == "" {
		return "", fmt.Errorf("env GITLAB_USER_LOGIN not set")
	}

	return actor, nil
}

// getInputForAction gets input to action
func getInputForAction(key string) string {
	return os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(key)))
}

// GetWorkDirectory gets workdir
func (p *Provider) GetWorkDirectory() string {
	workdirInput := getInputForAction("workdir")
	if workdirInput != "" {
		return workdirInput
	}

	return os.Getenv("CI_PROJECT_DIR")
}

// GetTemplateFile returns the template file
func (p *Provider) GetTemplateFile() string {
	templateFile := getInputForAction("krew_template_file")
	if templateFile != "" {
		return filepath.Join(p.GetWorkDirectory(), templateFile)
	}

	return filepath.Join(p.GetWorkDirectory(), ".krew.yaml")
}
