package actions

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

//Inputs is action inputs
type Inputs struct {
	PluginName                string
	Token                     string
	TokenUserHandle           string
	TokenUserEmail            string
	TokenUserName             string
	UpstreamKrewIndexRepoName string
	UpstreamKrewIndexOwner    string
}

//Derived is derived data
type Derived struct {
	UpstreamCloneURL string
	LocalCloneURL    string
}

//ActionData is action data
type ActionData struct {
	Workspace   string
	Actor       string
	Repo        string
	RepoOwner   string
	Inputs      Inputs
	ReleaseInfo *github.RepositoryRelease
	Derived     Derived
}

//Action defines interface to get data from actions
type Action interface {
	GetActionData(event *github.ReleaseEvent) ActionData
}

//RealAction is the real action
type RealAction struct {
	Token           string `yaml:"token"`
	TokenUserHandle string `yaml:"token-user-handle"`
	TokenUsername   string `yaml:"token-username"`
	TokenEmail      string `yaml:"token-email"`
	WebhookSecret   string `yaml:"webhook-secret"`
}

//GetActionData returns action data
func (r RealAction) GetActionData(event *github.ReleaseEvent) (ActionData, error) {
	if event.Release == nil {
		return ActionData{}, fmt.Errorf("event.Release is nil %v", event)
	}

	if len(event.Release.Assets) == 0 {
		return ActionData{}, fmt.Errorf("no release assets found")
	}

	//TODO: add validation that its indeed sent from a legitimate user
	releaseInfo := event.Release

	upstreamKrewIndexRepoName := r.getInputForAction("upstream-krew-index-repo-name")
	if upstreamKrewIndexRepoName == "" {
		upstreamKrewIndexRepoName = "krew-index"
	}

	upstreamKrewIndexRepoOwner := r.getInputForAction("upstream-krew-index-owner")
	if upstreamKrewIndexRepoOwner == "" {
		upstreamKrewIndexRepoOwner = "kubernetes-sigs"
	}

	return ActionData{
		Actor:       event.GetSender().GetLogin(),
		Repo:        event.GetRepo().GetName(),
		RepoOwner:   event.GetRepo().GetOwner().GetLogin(),
		ReleaseInfo: releaseInfo,
		Inputs: Inputs{
			UpstreamKrewIndexOwner:    upstreamKrewIndexRepoOwner,
			UpstreamKrewIndexRepoName: upstreamKrewIndexRepoName,
			Token:                     r.Token,
			TokenUserEmail:            r.TokenEmail,
			TokenUserName:             r.TokenUserHandle,
			TokenUserHandle:           r.TokenUserHandle,
		},
		Derived: Derived{
			UpstreamCloneURL: getRepoURL(upstreamKrewIndexRepoOwner, upstreamKrewIndexRepoName),
			LocalCloneURL:    getRepoURL("krew-release-bot", "krew-index"),
		},
	}, nil
}

func getRepoURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

//getInputForAction gets input to action
func (r RealAction) getInputForAction(key string) string {
	return os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(key)))
}

//GetPayload reads payload and returns it
func (r RealAction) getPayload() ([]byte, error) {
	eventJSONPath := os.Getenv("GITHUB_EVENT_PATH")
	data, err := ioutil.ReadFile(eventJSONPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
