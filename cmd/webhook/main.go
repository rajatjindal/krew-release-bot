package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/sirupsen/logrus"
)

func main() {
	ghToken := os.Getenv("GH_TOKEN")
	releaser, err := releaser.New(ghToken)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	lambda.Start(releaser.HandleActionLambdaWebhook)
}
