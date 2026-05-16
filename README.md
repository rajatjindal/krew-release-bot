[![Netlify Status](https://api.netlify.com/api/v1/badges/cfd72dea-e22a-463b-8e20-5748b743140a/deploy-status)](https://app.netlify.com/sites/angry-borg-f9dd47/deploys)

<a href="https://github.com/rajatjindal/krew-release-bot"><img src="https://github.com/krew-release-bot.png" width="100"></a><span width="10px">

`krew-release-bot` is a bot that automates the update of plugin manifests in `krew-index` when a new version of your `kubectl` plugin is released.
If a release is marked as a 'prerelease' in github, it will not be released to the krew index.

To trigger `krew-release-bot` you can use a `github-action` which sends the event to the bot.
If you provide a dedicated token for the target index repo, the action can also open the PR directly from CI without using the webhook service.

# Basic Setup

- Make sure you have enabled github actions for your repo
- Add a `.krew.yaml` template file at the root of your repo. Refer to [kubectl-evict-pod](https://github.com/rajatjindal/kubectl-evict-pod) repo for an example.
  - you could use https://rajatjindal.com/tools/krew-release-bot-helper/ for generating template for your plugin
- To setup the action, add the following snippet after the step that publishes the new release and assets:
  ```yaml
  - name: Update new version in krew-index
    uses: rajatjindal/krew-release-bot@v0.0.50
```
  Check out the `goreleaser` example below for details.

##### Example when using go-releaser

`<your-git-root>/.github/workflows/release.yml`

```yaml
name: release
on:
  push:
    tags:
      - "v*.*.*"
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@master
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.20'
      - name: GoReleaser
        uses: goreleaser/goreleaser-action@v1
        with:
          version: latest
          args: release --rm-dist
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      - name: Update new version in krew-index
        uses: rajatjindal/krew-release-bot@v0.0.50
```

\*\* You can also customize the release assets names, platforms for which build is done using .goreleaser.yml file in root of your git repo.

# Examples using krew-release-bot in different ways

- [bash based plugins](https://github.com/ahmetb/kubectx/blob/master/.github/workflows/release.yml)
- [multiple plugins published from one repo](https://github.com/ahmetb/kubectx/blob/master/.github/workflows/release.yml)
- [circle-ci](examples/circleci.yml)
- [travis-ci](examples/travis.yml)

# Testing the template file

You can test the template file rendering before check-in to the repo by running following command

```bash
$ docker run -v /path/to/your/template-file.yaml:/tmp/template-file.yaml ghcr.io/rajatjindal/krew-release-bot:v0.0.50 \
  krew-release-bot template --tag <tag-name> --template-file /tmp/template-file.yaml
```

# Inputs for the action

| Key                | Default Value          | Description                                                                          |
| ------------------ | ---------------------- | ------------------------------------------------------------------------------------ |
| workdir            | `env.GITHUB_WORKSPACE` | Overrides the GitHub workspace directory path                                        |
| krew_template_file | `.krew.yaml`           | The path to template file relative to $workdir. e.g. templates/misc/plugin-name.yaml |
| index_repo_token   | empty                  | Optional token with write access to the target index repo. Enables direct PR creation from CI |
| index_repo_provider | `github`             | Optional git hosting provider for the target index repo                               |
| index_pr_provider  | `index_repo_provider` | Optional pull request provider for the target index repo                              |
| upstream_krew_index_repo_owner   | `kubernetes-sigs`     | Optional owner for the upstream krew index repo                                       |
| upstream_krew_index_repo_name    | `krew-index`          | Optional name for the upstream krew index repo                                        |
| upstream_krew_index_repo_clone_url | provider default    | Optional clone URL override for the upstream krew index repo                          |

# Direct PR mode for custom krew index repos

If you want to open PRs against your own index repo instead of `kubernetes-sigs/krew-index`, provide a token with push and pull request permissions to that repo and set the target repo inputs.

```yaml
- name: Update new version in custom krew-index
  uses: rajatjindal/krew-release-bot@v0.0.50
  with:
    index_repo_token: ${{ secrets.KREW_INDEX_TOKEN }}
    index_repo_provider: github
    index_pr_provider: github
    upstream_krew_index_repo_owner: your-org
    upstream_krew_index_repo_name: custom-krew-index
```

When `index_repo_token` is set, the action skips the webhook service, pushes a release branch directly to the configured index repo, and opens the PR from CI.

The config surface is provider-neutral so git hosting and PR backends can evolve independently. At the moment, the implemented direct git and PR provider is `github`.

# Limitations of krew-release-bot

- only works for repos hosted on github right now
- The first version of plugin has to be submitted manually, by plugin author, to the krew-index repo

# Kubernetes CLA

krew-release-bot is just a service to open PR on your behalf to release a new version of the krew-plugin. Your CLA agreement (that you did when submitting the new plugin to krew-index) is still applicable on these PR's.
