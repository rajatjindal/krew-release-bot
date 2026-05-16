package releaser

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rajatjindal/krew-release-bot/pkg/types"
)

// GithubActions is github webhook handler
type GithubActions struct{}

// NewGithubActions gets new git webhook instance
func newGithubActions() (*GithubActions, error) {
	return &GithubActions{}, nil
}

// Parse validates the request
func (w *GithubActions) Parse(r *http.Request) (*types.ReleaseRequest, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &types.ReleaseRequest{}
	err = json.Unmarshal(body, request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

// ParseLambdaRequest parses the request from lambda request object
func (w *GithubActions) ParseLambdaRequest(r events.APIGatewayProxyRequest) (*types.ReleaseRequest, error) {
	request := &types.ReleaseRequest{}
	err := json.Unmarshal([]byte(r.Body), request)
	if err != nil {
		return nil, err
	}

	return request, nil
}
