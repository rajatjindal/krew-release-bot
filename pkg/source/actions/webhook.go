package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
)

//GithubActions is github webhook handler
type GithubActions struct{}

//NewGithubActions gets new git webhook instance
func NewGithubActions() (*GithubActions, error) {
	return &GithubActions{}, nil
}

//Parse validates the request
func (w *GithubActions) Parse(r *http.Request) (*source.ReleaseRequest, error) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &source.ReleaseRequest{}
	err = json.Unmarshal(body, request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

type authInjector struct {
	token string
}

func (ij *authInjector) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", ij.token))
	return http.DefaultTransport.RoundTrip(req)
}

func validateTokenAndFetchRequest(r *http.Request) (*source.ReleaseRequest, error) {
	token := r.Header.Get("x-github-token")
	if token == "" {
		return nil, fmt.Errorf("token is empty")
	}

	mc := &http.Client{
		Transport: &authInjector{token: token},
	}
	client := github.NewClient(mc)

	repos, _, err := client.Apps.ListRepos(context.TODO(), &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repos can be accessed using this token")
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &source.ReleaseRequest{}
	err = json.Unmarshal(body, request)
	if err != nil {
		return nil, err
	}

	// This check is to ensure no one fakes the request to our webhook
	// without this, the attacker can potentially claim to submit a release request
	// for e.g. kubectl-whoami plugin when he is not authorized to do so
	hasAccess := false
	for _, repo := range repos {
		if repo.GetOwner().GetLogin() == request.PluginOwner && repo.GetName() == request.PluginRepo {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, fmt.Errorf("provided token do not have access to repo %s/%s", request.PluginOwner, request.PluginRepo)
	}

	return request, nil
}
