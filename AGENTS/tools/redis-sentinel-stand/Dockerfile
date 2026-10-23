# Minimal runtime for the Pro binary cross-compiled by stand.sh build (linux, CGO_ENABLED=0).
FROM alpine:3.24
RUN apk add --no-cache bash git curl ca-certificates openssh-client
COPY bin/semaphore /usr/local/bin/semaphore
COPY entrypoint.sh /entrypoint.sh
COPY repo /repo-src
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
