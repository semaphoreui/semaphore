FROM alpine:3.15 AS builder
ENV GOBIN="/usr/local/bin"
COPY ./ /semaphore
WORKDIR /semaphore
RUN apk --no-cache -U add curl git go libc-dev nodejs npm python3
RUN go install github.com/go-task/task/v3/cmd/task@latest
RUN ./deployment/docker/prod/bin/install


FROM alpine:3.15 AS release
ENV SEMAPHORE_ARCH="linux_amd64" \
    SEMAPHORE_CONFIG_PATH="${SEMAPHORE_CONFIG_PATH:-/etc/semaphore}" \
    SEMAPHORE_VERSION="production"

RUN apk --no-cache add ansible-core ca-certificates curl git gcompat mysql-client nodejs npm openssh-client openssl py3-netaddr py3-openssl py3-pip python3 rsync sshpass tini tzdata

RUN addgroup semaphore && adduser -D -G semaphore semaphore
RUN mkdir -p /etc/semaphore /tmp/semaphore
RUN chown -R semaphore /etc/semaphore /tmp/semaphore

COPY --chown=semaphore --from=builder /usr/local/bin/semaphore-wrapper /usr/local/bin/
COPY --chown=semaphore --from=builder /usr/local/bin/semaphore /usr/local/bin/

WORKDIR /home/semaphore
USER semaphore

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/semaphore-wrapper", "/usr/local/bin/semaphore", "server", "--config", "/etc/semaphore/config.json"]
