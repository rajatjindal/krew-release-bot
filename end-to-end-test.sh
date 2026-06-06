export GITHUB_ACTIONS=true
export GITHUB_TOKEN=ghp_your_pat_here
export GITHUB_REPOSITORY=rajatjindal/kubectl-whoami
export GITHUB_ACTOR=rajatjindal
export GITHUB_WORKSPACE="$(pwd)"

export INPUT_WORKDIR="$(pwd)"
export INPUT_KREW_TEMPLATE_FILE=.krew.yaml
export INPUT_KREW_PLUGIN_RELEASE_TAG=v0.0.48
export INPUT_SUBMIT_PR_LOCALLY=true
export INPUT_DRY_RUN=true

export INPUT_UPSTREAM_KREW_INDEX_REPO_URL=https://github.com/rajatjin/krew-index.git
export INPUT_LOCAL_KREW_INDEX_REPO_URL=https://github.com/rajatjin/krew-index.git

./krew-release-bot action