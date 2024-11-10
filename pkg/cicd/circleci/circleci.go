package circleci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Provider implements provider interface
type Provider struct{}

// GetTag returns tag
func (p *Provider) GetTag() (string, error) {
	ref := os.Getenv("CIRCLE_TAG")
	if ref == "" {
		return "", fmt.Errorf("CIRCLE_TAG env variable not found")
	}

	return ref, nil
}

// GetOwnerAndRepo gets the owner and repo from the env
func (p *Provider) GetOwnerAndRepo() (string, string, error) {
	owner := os.Getenv("CIRCLE_PROJECT_USERNAME")
	if owner == "" {
		return "", "", fmt.Errorf("env CIRCLE_PROJECT_USERNAME not set")
	}

	repo := os.Getenv("CIRCLE_PROJECT_REPONAME")
	if repo == "" {
		return "", "", fmt.Errorf("env CIRCLE_PROJECT_REPONAME not set")
	}

	return owner, repo, nil
}

// GetActor gets the owner and repo from the env
func (p *Provider) GetActor() (string, error) {
	actor := os.Getenv("CIRCLE_USERNAME")
	if actor == "" {
		return "", fmt.Errorf("env CIRCLE_USERNAME not set")
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

	dir := os.Getenv("CIRCLE_WORKING_DIRECTORY")

	//workaround for https://discuss.circleci.com/t/circle-working-directory-doesnt-expand/17007/3
	if dir == "~/project" {
		return filepath.Join(os.Getenv("HOME"), "project")
	}

	return dir
}

// GetTemplateFile returns the template file
func (p *Provider) GetTemplateFile() string {
	templateFile := getInputForAction("krew_template_file")
	if templateFile != "" {
		return filepath.Join(p.GetWorkDirectory(), templateFile)
	}

	return filepath.Join(p.GetWorkDirectory(), ".krew.yaml")
}
