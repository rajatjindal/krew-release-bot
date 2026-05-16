package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/rajatjindal/krew-release-bot/pkg/webhook"
)

func main() {
	ghToken := os.Getenv("GH_TOKEN")
	releaser := releaser.New(ghToken)
	handler := webhook.NewActionHandler(releaser)

	lambda.Start(handler.HandleActionLambdaWebhook)
}
