package releaser

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
)

// Releaser is what opens PR
type Releaser struct {
	Token           string
	TokenEmail      string
	TokenUserHandle string
	TokenUsername   string
	Config          IndexRepoConfig

	forge Forge
}

type releaseRunner interface {
	Release(request *source.ReleaseRequest) (string, error)
}

var newReleaseRunnerFromConfig = func(config IndexRepoConfig) (releaseRunner, error) {
	return NewFromConfig(config)
}

func newReleaserFromConfig(config IndexRepoConfig) (*Releaser, error) {
	forge, err := NewForge(
		config.Upstream.ForgeKind,
		config.Upstream.APIBaseURL,
		config.Upstream.Auth.Token,
	)
	if err != nil {
		return nil, err
	}

	return newReleaserWithForge(forge, config)
}

func newReleaserWithForge(forge Forge, config IndexRepoConfig) (*Releaser, error) {
	currentUser, err := forge.CurrentUser()
	if err != nil {
		return nil, err
	}

	return &Releaser{
		Token:           config.LocalPushTarget.Auth.Token,
		TokenEmail:      currentUser.Email,
		TokenUserHandle: currentUser.Handle,
		TokenUsername:   currentUser.Name,
		Config:          config,
		forge:           forge,
	}, nil
}

func NewFromConfig(config IndexRepoConfig) (*Releaser, error) {
	return newReleaserFromConfig(config)
}

// HandleActionLambdaWebhook handles requests from github actions
func HandleActionLambdaWebhook(ctx context.Context, request events.APIGatewayProxyRequest, config IndexRepoConfig) (*events.APIGatewayProxyResponse, error) {
	releaser, err := newReleaseRunnerFromConfig(config)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       errors.Wrap(err, "failed to create releaser with the provided token").Error(),
		}, nil
	}

	hook, err := NewGithubActions()
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       errors.Wrap(err, "creating instance of action handler").Error(),
		}, nil
	}

	releaseRequest, err := hook.ParseLambdaRequest(request)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       errors.Wrap(err, "getting release request").Error(),
		}, nil
	}

	pr, err := releaser.Release(releaseRequest)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       errors.Wrap(err, "opening pr").Error(),
		}, nil
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       fmt.Sprintf("PR %q submitted successfully", pr),
	}, nil
}

// HandleActionWebhook handles requests from github actions
func HandleActionWebhook(w http.ResponseWriter, r *http.Request, config IndexRepoConfig) {
	releaser, err := newReleaseRunnerFromConfig(config)
	if err != nil {
		http.Error(w, errors.Wrap(err, "failed to create releaser with the provided token").Error(), http.StatusInternalServerError)
		return
	}

	hook, err := NewGithubActions()
	if err != nil {
		http.Error(w, errors.Wrap(err, "creating instance of action handler").Error(), http.StatusInternalServerError)
		return
	}

	releaseRequest, err := hook.Parse(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting release request").Error(), http.StatusInternalServerError)
		return
	}

	pr, err := releaser.Release(releaseRequest)
	if err != nil {
		http.Error(w, errors.Wrap(err, "opening pr").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("PR %q submitted successfully", pr)))
}
