package validator

import (
	"context"

	"github.com/rajatjindal/krew-release-bot/pkg/source/validator/github"
)

type Validator interface {
	Validate(ctx context.Context, owner, repo, tag string) error
}

func GetValidator() Validator {
	return &github.Validator{}
}
