package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pkg/errors"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/rajatjindal/krew-release-bot/pkg/source/actions"
)

type ActionHandler struct {
	releaser *releaser.Releaser
}

func NewActionHandler(r *releaser.Releaser) *ActionHandler {
	return &ActionHandler{releaser: r}
}

// HandleActionLambdaWebhook handles requests from github actions.
func (h *ActionHandler) HandleActionLambdaWebhook(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	hook, err := actions.NewGithubActions()
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

	pr, err := h.releaser.Release(releaseRequest)
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

// HandleActionWebhook handles requests from github actions.
func (h *ActionHandler) HandleActionWebhook(w http.ResponseWriter, r *http.Request) {
	hook, err := actions.NewGithubActions()
	if err != nil {
		http.Error(w, errors.Wrap(err, "creating instance of action handler").Error(), http.StatusInternalServerError)
		return
	}

	releaseRequest, err := hook.Parse(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting release request").Error(), http.StatusInternalServerError)
		return
	}

	pr, err := h.releaser.Release(releaseRequest)
	if err != nil {
		http.Error(w, errors.Wrap(err, "opening pr").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("PR %q submitted successfully", pr)))
}
