package github

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Validator struct{}

func getHTTPClient(ctx context.Context) *http.Client {
	if os.Getenv("GITHUB_TOKEN") != "" {
		logrus.Info("GITHUB_TOKEN env variable found, using authenticated requests.")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
		return oauth2.NewClient(ctx, ts)
	}

	return nil
}

func (r *Validator) Validate(ctx context.Context, owner, repo, tag string) error {
	client := github.NewClient(getHTTPClient(ctx))

	release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		return err
	}

	if release.GetPrerelease() {
		return fmt.Errorf("release with tag %q is a pre-release. skipping", release.GetTagName())
	}

	return nil
}
