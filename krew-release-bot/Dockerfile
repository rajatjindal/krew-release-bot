FROM golang:1.13.4-alpine3.10 as builder

WORKDIR /go/src/github.com/rajatjindal/krew-release-bot/krew-release-bot
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -o krew-release-bot main.go

FROM openfaas/classic-watchdog:0.18.1 as watchdog

FROM alpine:3.10.3

RUN mkdir -p /home/app

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

# Add non root user
RUN addgroup -S app && adduser app -S -G app
RUN chown app /home/app

WORKDIR /home/app

USER app

COPY --from=builder /go/src/github.com/rajatjindal/krew-release-bot/krew-release-bot/krew-release-bot /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/krew-release-bot"]

# Populate example here - i.e. "cat", "sha512sum" or "node index.js"
ENV fprocess="/usr/local/bin/krew-release-bot"
# Set to true to see request in function logs
ENV write_debug="false"

EXPOSE 8080

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]