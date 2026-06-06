package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
)

func main() {
	config := getIndexRepoConfigFromEnv()
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		return releaser.HandleActionLambdaWebhook(ctx, request, config)
	})
}
