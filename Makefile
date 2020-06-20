.PHONY: lambda

lambda:
	mkdir -p functions
	CGO_ENABLED=0 GOOS=linux go build -mod vendor --ldflags "-s -w" -o functions/github-action-webhook cmd/webhook/main.go

all: lambda