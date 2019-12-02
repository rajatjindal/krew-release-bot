<a href="https://github.com/rajatjindal/krew-release-bot"><img src="https://github.com/krew-plugin-release-bot.png" width="100"></a><span width="10px">

krew-release-bot is a bot that listens for release events from krew plugins repos which have configured the webhooks for listening to release event.

On release event it run few validations and then open the PR against [kubernetes-sigs/krew-index](https://kubernetes-sigs/krew-index) repo to release new version of your plugin.

# Setup

- Install [krew-release-bot](https://github.com/apps/krew-release-bot) github app on the repo(s)
- Add a `.krew.yaml` template file at the root of your repo. Refer to [kubectl-whoami](https://github.com/rajatjindal/kubectl-whoami) repo for an example.
- Publish a new release version of your plugin
- The bot will use `.krew.yaml` template and generate the plugin spec file for your plugin and open the PR for krew-index

# Limitations
- only works for repos hosted on github right now
- only supports one plugin per git repo right now
- The first version of plugin has to be submitted manually, by plugin author, to the krew-index repo
- The homepage in the plugin spec in krew-index is used to establish ownership. The repo from which the release is published should be the homepage of the plugin in already released plugin-spec.

# How to Install

- Go to https://github.com/apps/krew-release-bot
- Click on Configure
- Select the User/Org which owns the repo where you plan to install this app.
- Confirm Password (required by `github`). App don't get access to this password.
- Refer that `read` access is required to `code` and `metadata` to listen to `release` events.
- From `Repository Access` box, select the repositories where you want to enable it. You can enable for `all` or `only selected` repositories.
- Click Save and you are all set.

# Permissions required

The github app needs `read` access to `code` and `metadata` of the repository. Refer to the screenshot below:

![Permissions](docs/permissions.png)

# Kubernetes CLA

krew-release-bot is just a service to open PR on your behalf to release a new version of the krew-plugin. Your CLA agreement (that you did when submitting the new plugin to krew-index) is still applicable on these PR's. 
