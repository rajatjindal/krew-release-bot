recommended way to use krew-release-bot is through github-actions. look at readme at root of this repo for example.

### Configuration when using github app

#### How to Install

- Go to [`https://github.com/apps/krew-release-bot`](https://github.com/apps/krew-release-bot)
- Click on Configure
- Select the User/Org which owns the repo where you plan to install this app.
- Confirm Password (required by `github`). App don't get access to this password.
- Refer that `read` access is required to `code` and `metadata` to listen to `release` events.
- From `Repository Access` box, select the repositories where you want to enable it. You can enable for `all` or `only selected` repositories.
- Click Save and you are all set.

#### Permissions required

The github app needs `read` access to `code` and `metadata` of the repository. Refer to the screenshot below:

![Permissions](permissions.png)
