package travisci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Provider implements provider interface
type Provider struct{}

// TODO(rajatjindal): implement for travisci
func (p *Provider) IsPreRelease(owner, repo, tag string) (bool, error) {
	return false, nil
}

// GetTag returns tag
func (p *Provider) GetTag() (string, error) {
	ref := os.Getenv("TRAVIS_TAG")
	if ref == "" {
		return "", fmt.Errorf("TRAVIS_TAG env variable not found")
	}

	return ref, nil
}

// GetOwnerAndRepo gets the owner and repo from the env
func (p *Provider) GetOwnerAndRepo() (string, string, error) {
	repoFromEnv := os.Getenv("TRAVIS_REPO_SLUG")
	if repoFromEnv == "" {
		return "", "", fmt.Errorf("env TRAVIS_REPO_SLUG not set")
	}

	s := strings.Split(repoFromEnv, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("env TRAVIS_REPO_SLUG is incorrect format. expected format <owner>/<repo>, found %q", repoFromEnv)
	}

	return s[0], s[1], nil
}

// GetActor gets the owner and repo from the env
func (p *Provider) GetActor() (string, error) {
	owner, _, err := p.GetOwnerAndRepo()
	if err != nil {
		return "", err
	}

	if owner == "" {
		return "", fmt.Errorf("failed to find actor for the release")
	}

	return owner, nil
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

	return os.Getenv("TRAVIS_BUILD_DIR")
}

// GetTemplateFile returns the template file
func (p *Provider) GetTemplateFile() string {
	templateFile := getInputForAction("krew_template_file")
	if templateFile != "" {
		return filepath.Join(p.GetWorkDirectory(), templateFile)
	}

	return filepath.Join(p.GetWorkDirectory(), ".krew.yaml")
}

// GetKrewIndexRepoName gets upstream_krew_index_repo_name
func (p *Provider) GetKrewIndexRepoName() string {
	nameInput := getInputForAction("UPSTREAM_KREW_INDEX_REPO_NAME")
	if nameInput != "" {
		return nameInput
	}

	return os.Getenv("UPSTREAM_KREW_INDEX_REPO_NAME")
}

// GetKrewIndexRepoName gets upstream_krew_index_repo_owner
func (p *Provider) GetKrewIndexRepoOwner() string {
	ownerInput := getInputForAction("UPSTREAM_KREW_INDEX_REPO_OWNER")
	if ownerInput != "" {
		return ownerInput
	}

	return os.Getenv("UPSTREAM_KREW_INDEX_REPO_OWNER")
}
