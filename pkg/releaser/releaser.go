package releaser

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
	"github.com/rajatjindal/krew-release-bot/pkg/krew"
)

// Releaser is what opens PR
type Releaser struct {
	Token                         string
	TokenEmail                    string
	TokenUserHandle               string
	TokenUsername                 string
	UpstreamKrewIndexRepo         string
	UpstreamKrewIndexRepoOwner    string
	UpstreamKrewIndexRepoCloneURL string
	LocalKrewIndexRepo            string
	LocalKrewIndexRepoOwner       string
	LocalKrewIndexRepoCloneURL    string
}

// New returns new releaser object
func New(ghToken string) (*Releaser, error) {
	tokenUserHandle, tokenUsername, tokenEmail, err := getUserDetails(ghToken)
	if err != nil {
		return nil, fmt.Errorf("could not derive user information from the github token")
	}

	return &Releaser{
		Token:                         ghToken,
		TokenEmail:                    tokenEmail,
		TokenUserHandle:               tokenUserHandle,
		TokenUsername:                 tokenUsername,
		UpstreamKrewIndexRepo:         krew.GetUpstreamKrewIndexRepoName(),
		UpstreamKrewIndexRepoOwner:    krew.GetUpstreamKrewIndexRepoOwner(),
		UpstreamKrewIndexRepoCloneURL: getCloneURL(krew.GetUpstreamKrewIndexRepoOwner(), krew.GetUpstreamKrewIndexRepoName()),
		LocalKrewIndexRepo:            getLocalKrewIndexRepoName(),
		LocalKrewIndexRepoOwner:       tokenUserHandle,
		LocalKrewIndexRepoCloneURL:    getLocalKrewIndexRepo(tokenUserHandle, getLocalKrewIndexRepoName()),
	}, nil
}

// HandleActionLambdaWebhook handles requests from github actions
func (releaser *Releaser) HandleActionLambdaWebhook(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	hook, err := newGithubActions()
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
func (releaser *Releaser) HandleActionWebhook(w http.ResponseWriter, r *http.Request) {
	hook, err := newGithubActions()
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
