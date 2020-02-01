package githubactions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//Actions implements provider interface
type Actions struct{}

//GetTag returns tag
func (p *Actions) GetTag() (string, error) {
	ref := os.Getenv("GITHUB_REF")
	if ref == "" {
		return "", fmt.Errorf("GITHUB_REF env variable not found")
	}

	//GITHUB_REF=refs/tags/v0.0.6
	if !strings.HasPrefix(ref, "refs/tags/") {
		return "", fmt.Errorf("GITHUB_REF expected to be of format refs/tags/<tag> but found %q", ref)
	}

	return strings.ReplaceAll(ref, "refs/tags/", ""), nil
}

//GetOwnerAndRepo gets the owner and repo from the env
func (p *Actions) GetOwnerAndRepo() (string, string, error) {
	repoFromEnv := os.Getenv("GITHUB_REPOSITORY")
	if repoFromEnv == "" {
		return "", "", fmt.Errorf("env GITHUB_REPOSITORY not set")
	}

	s := strings.Split(repoFromEnv, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("env GITHUB_REPOSITORY is incorrect format. expected format <owner>/<repo>, found %q", repoFromEnv)
	}

	return s[0], s[1], nil
}

//GetActor gets the owner and repo from the env
func (p *Actions) GetActor() (string, error) {
	actor := os.Getenv("GITHUB_ACTOR")
	if actor == "" {
		return "", fmt.Errorf("env GITHUB_ACTOR not set")
	}

	return actor, nil
}

//getInputForAction gets input to action
func getInputForAction(key string) string {
	return os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(key)))
}

//GetWorkDirectory gets workdir
func (p *Actions) GetWorkDirectory() string {
	workdirInput := getInputForAction("workdir")
	if workdirInput != "" {
		return workdirInput
	}

	return os.Getenv("GITHUB_WORKSPACE")
}

//GetTemplateFile returns the template file
func (p *Actions) GetTemplateFile() string {
	templateFile := getInputForAction("krew_template_file")
	if templateFile != "" {
		return filepath.Join(p.GetWorkDirectory(), templateFile)
	}

	return filepath.Join(p.GetWorkDirectory(), ".krew.yaml")
}
