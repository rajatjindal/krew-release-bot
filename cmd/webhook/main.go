package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/rajatjindal/krew-release-bot/pkg/webhook"
)

func main() {
	ghToken := os.Getenv("GH_TOKEN")
	releaser, err := releaser.New(releaser.ProviderGitHub, ghToken)
	if err != nil {
		panic(err)
	}
	handler := webhook.NewActionHandler(releaser)

	lambda.Start(handler.HandleActionLambdaWebhook)
}
