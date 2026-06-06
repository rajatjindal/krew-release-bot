package releaser

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rajatjindal/krew-release-bot/pkg/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubReleaseRunner struct {
	releaseFunc func(request *source.ReleaseRequest) (string, error)
}

func (s *stubReleaseRunner) Release(request *source.ReleaseRequest) (string, error) {
	return s.releaseFunc(request)
}

func TestHandleActionWebhookHappyPath(t *testing.T) {
	originalFactory := newReleaseRunnerFromConfig
	defer func() { newReleaseRunnerFromConfig = originalFactory }()

	config := IndexRepoConfigFromRaw(RawIndexRepoConfig{
		Upstream: RawReleaseTarget{
			ForgeKind: ForgeKindGitHub,
			RepoURL:   "https://github.com/kubernetes-sigs/krew-index.git",
		},
		LocalPushTarget: RawReleaseTarget{
			ForgeKind: ForgeKindGitHub,
			RepoURL:   "https://github.com/krew-release-bot/krew-index.git",
		},
	})

	var capturedConfig IndexRepoConfig
	var capturedRequest *source.ReleaseRequest
	newReleaseRunnerFromConfig = func(actual IndexRepoConfig) (releaseRunner, error) {
		capturedConfig = actual
		return &stubReleaseRunner{
			releaseFunc: func(request *source.ReleaseRequest) (string, error) {
				capturedRequest = request
				return "https://example.test/pr/1", nil
			},
		}, nil
	}

	payload := &source.ReleaseRequest{
		TagName:            "v1.2.3",
		PluginName:         "whoami",
		PluginRepo:         "kubectl-whoami",
		PluginOwner:        "rajatjindal",
		PluginReleaseActor: "release-user",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/github-action-webhook", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	HandleActionWebhook(recorder, req, config)

	require.NotNil(t, capturedRequest)
	assert.Equal(t, config, capturedConfig)
	assert.Equal(t, payload, capturedRequest)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `PR "https://example.test/pr/1" submitted successfully`)
}

func TestHandleActionLambdaWebhookHappyPath(t *testing.T) {
	originalFactory := newReleaseRunnerFromConfig
	defer func() { newReleaseRunnerFromConfig = originalFactory }()

	config := IndexRepoConfigFromRaw(RawIndexRepoConfig{
		Upstream: RawReleaseTarget{
			ForgeKind: ForgeKindGitHub,
			RepoURL:   "https://github.com/kubernetes-sigs/krew-index.git",
		},
		LocalPushTarget: RawReleaseTarget{
			ForgeKind: ForgeKindGitHub,
			RepoURL:   "https://github.com/krew-release-bot/krew-index.git",
		},
	})

	var capturedRequest *source.ReleaseRequest
	newReleaseRunnerFromConfig = func(actual IndexRepoConfig) (releaseRunner, error) {
		assert.Equal(t, config, actual)
		return &stubReleaseRunner{
			releaseFunc: func(request *source.ReleaseRequest) (string, error) {
				capturedRequest = request
				return "https://example.test/pr/2", nil
			},
		}, nil
	}

	payload := &source.ReleaseRequest{
		TagName:            "v2.0.0",
		PluginName:         "whoami",
		PluginRepo:         "kubectl-whoami",
		PluginOwner:        "rajatjindal",
		PluginReleaseActor: "release-user",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	response, err := HandleActionLambdaWebhook(context.Background(), events.APIGatewayProxyRequest{
		Body: string(body),
	}, config)
	require.NoError(t, err)

	require.NotNil(t, capturedRequest)
	assert.Equal(t, payload, capturedRequest)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Contains(t, response.Body, `PR "https://example.test/pr/2" submitted successfully`)
}
