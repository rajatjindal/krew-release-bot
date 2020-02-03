<a href="https://github.com/rajatjindal/krew-release-bot"><img src="https://github.com/krew-release-bot.png" width="100"></a><span width="10px">

`krew-release-bot` is a bot that automates the update of plugin manifests in `krew-index` when a new version of your `kubectl` plugin is released.

To trigger `krew-release-bot` you can use a `github-action` which sends the event to the bot.

# Basic Setup
- Make sure you have enabled github actions for your repo
- Add a `.krew.yaml` template file at the root of your repo. Refer to [kubectl-evict-pod](https://github.com/rajatjindal/kubectl-evict-pod) repo for an example.
- To setup the action, add the following snippet after the step that publishes the new release and assets:
  ```yaml
  - name: Update new version in krew-index
    uses: rajatjindal/krew-release-bot@v0.0.36
  ```
  Check out the `goreleaser` example below for details.

##### Example when using go-releaser

`<your-git-root>/.github/workflows/release.yml`

```yaml
name: release
on:
  push:
    tags:
    - 'v*.*.*'
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@master
    - name: Setup Go
      uses: actions/setup-go@v1
      with:
        go-version: 1.13
    - name: GoReleaser
      uses: goreleaser/goreleaser-action@v1
      with:
        version: latest
        args: release --rm-dist
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    - name: Update new version in krew-index
      uses: rajatjindal/krew-release-bot@v0.0.36
```

** You can also customize the release assets names, platforms for which build is done using .goreleaser.yml file in root of your git repo.

# Examples using krew-release-bot in different ways

- [bash based plugins](https://github.com/ahmetb/kubectx/blob/master/.github/workflows/release.yml)
- [multiple plugins published from one repo](https://github.com/ahmetb/kubectx/blob/master/.github/workflows/release.yml)
- [circle-ci](examples/circleci.yml)
- [travis-ci](examples/travis.yml)

# Testing the template file

You can test the template file rendering before check-in to the repo by running following command
```bash
$ docker run -v /path/to/your/template-file.yaml:/tmp/template-file.yaml rajatjindal/krew-release-bot:v0.0.36 \
  krew-release-bot template --tag <tag-name> --template-file /tmp/template-file.yaml
```

# Inputs for the action

| Key           | Default Value | Description |
| ------------- | ------------- | ----------- |
| workdir     | `env.GITHUB_WORKSPACE`  | Overrides the GitHub workspace directory path |
| krew_template_file  | `.krew.yaml`  | The path to template file relative to $workdir. e.g. templates/misc/plugin-name.yaml |


# Limitations of krew-release-bot
- only works for repos hosted on github right now
- only supports one plugin per git repo right now
- The first version of plugin has to be submitted manually, by plugin author, to the krew-index repo
- The homepage in the plugin spec in krew-index is used to establish ownership. The repo from which the release is published should be the homepage of the plugin in already released plugin-spec.


# Kubernetes CLA

krew-release-bot is just a service to open PR on your behalf to release a new version of the krew-plugin. Your CLA agreement (that you did when submitting the new plugin to krew-index) is still applicable on these PR's. 
