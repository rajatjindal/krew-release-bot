package release

import (
	"context"
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/release/github"
	"github.com/rajatjindal/krew-release-bot/pkg/release/gitlab"
)

type Validator interface {
	Validate(ctx context.Context, owner, repo, tag string) error
}

func GetValidator() (Validator, error) {
	if os.Getenv("GITLAB_CI") == "true" {
		return &gitlab.Validator{}, nil
	}

	return &github.Validator{}, nil
}
