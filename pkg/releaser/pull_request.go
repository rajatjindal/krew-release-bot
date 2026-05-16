package releaser

import (
	"fmt"

	"github.com/rajatjindal/krew-release-bot/pkg/source"
)

// PullRequest describes the data needed by a PR backend.
type PullRequest struct {
	Title string
	Head  string
	Base  string
	Body  string
}

// PullRequestOpener opens pull requests against a remote provider.
type PullRequestOpener interface {
	Open(owner, repo string, pullRequest PullRequest) (string, error)
}

func (r *Releaser) buildPullRequest(request *source.ReleaseRequest) PullRequest {
	return PullRequest{
		Title: r.getTitle(request),
		Head:  r.getHead(request),
		Base:  r.UpstreamKrewIndexBaseBranch,
		Body:  r.getPRBody(request),
	}
}

func (r *Releaser) getTitle(request *source.ReleaseRequest) string {
	return fmt.Sprintf(
		"release new version %s of %s",
		request.TagName,
		request.PluginName,
	)
}

func (r *Releaser) getBranchName(request *source.ReleaseRequest) string {
	return fmt.Sprintf("%s-%s-%s-%s", request.PluginOwner, request.PluginName, request.PluginRepo, request.TagName)
}

func (r *Releaser) getHead(request *source.ReleaseRequest) string {
	branchName := r.getBranchName(request)
	if r.RepositoryProvider == nil {
		return fmt.Sprintf("%s:%s", r.TokenUserHandle, branchName)
	}

	return r.RepositoryProvider.FormatPullRequestHead(PullRequestHeadInput{
		BranchName:        branchName,
		LocalRepoOwner:    r.LocalKrewIndexRepoOwner,
		LocalRepoName:     r.LocalKrewIndexRepo,
		UpstreamRepoOwner: r.UpstreamKrewIndexRepoOwner,
		UpstreamRepoName:  r.UpstreamKrewIndexRepo,
		TokenUserHandle:   r.TokenUserHandle,
	})
}

func (r *Releaser) getPRBody(request *source.ReleaseRequest) string {
	prBody := `hey krew-index team,

I am [krew-release-bot](https://github.com/rajatjindal/krew-release-bot), and I would like to open this PR to publish version %s of %s on behalf of @%s.

Thanks,
@krew-release-bot`

	return fmt.Sprintf(prBody,
		fmt.Sprintf("`%s`", request.TagName),
		fmt.Sprintf("`%s`", request.PluginName),
		request.PluginReleaseActor,
	)
}
