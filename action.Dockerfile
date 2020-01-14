FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /go/src/github.com/rajatjindal/krew-release-bot
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor --ldflags "-s -w" -o krew-release-bot cmd/action/main.go

FROM alpine:3.10.3

RUN mkdir -p /home/app

# Add non root user
RUN addgroup -S app && adduser app -S -G app
RUN chown app /home/app

WORKDIR /home/app

USER app

COPY --from=builder /go/src/github.com/rajatjindal/krew-release-bot/krew-release-bot /usr/local/bin/

ENTRYPOINT "/usr/local/bin/krew-release-bot"