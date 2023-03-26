package actions

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
)

// GithubActions is github webhook handler
type GithubActions struct{}

// NewGithubActions gets new git webhook instance
func NewGithubActions() (*GithubActions, error) {
	return &GithubActions{}, nil
}

// Parse validates the request
func (w *GithubActions) Parse(r *http.Request) (*source.ReleaseRequest, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
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

// ParseLambdaRequest parses the request from lambda request object
func (w *GithubActions) ParseLambdaRequest(r events.APIGatewayProxyRequest) (*source.ReleaseRequest, error) {
	request := &source.ReleaseRequest{}
	err := json.Unmarshal([]byte(r.Body), request)
	if err != nil {
		return nil, err
	}

	return request, nil
}
