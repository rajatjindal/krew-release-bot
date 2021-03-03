FROM golang:1.15.8-alpine3.13 as builder

WORKDIR /go/src/github.com/rajatjindal/krew-release-bot
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go test -mod vendor ./... -cover
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor --ldflags "-s -w" -o krew-release-bot cmd/action/*

FROM alpine:3.13.2

RUN mkdir -p /home/app

# Add non root user
RUN addgroup -S app && adduser app -S -G app
RUN chown app /home/app

WORKDIR /home/app

USER app

COPY --from=builder /go/src/github.com/rajatjindal/krew-release-bot/krew-release-bot /usr/local/bin/

CMD ["krew-release-bot", "action"]