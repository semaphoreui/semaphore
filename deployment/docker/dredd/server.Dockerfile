# Semaphore server for the Dredd API contract tests (task dredd:docker).
#
# The binary and the web UI are built from the working tree, exactly like the
# production image, but the image carries none of the Ansible / Terraform /
# OpenTofu runtimes: the contract tests only talk to the API and never run a
# task. That keeps a rebuild after a code change down to the Go/Vue compile.
#
# Test-only image: it runs as root so that the SQLite variant can share one
# database file with the dredd container. Never publish it.
FROM golang:1.26-alpine3.24 AS builder

RUN apk add --no-cache -U curl git nodejs npm

WORKDIR /usr/local
# hadolint ignore=DL4006
RUN curl -sL https://taskfile.dev/install.sh | sh

ENV GOWORK=off
WORKDIR /go/src/semaphore

# Dependencies first so they are cached until go.mod / package.json change.
# go.mod replaces the pro stub with ./pro, so its module files must be
# present for the download to resolve.
COPY go.mod go.sum ./
COPY pro/go.mod pro/go.sum ./pro/
RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

COPY web/package.json web/package-lock.json ./web/
RUN --mount=type=cache,target=/root/.npm \
    cd web && npm install

COPY . .

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task build

FROM alpine:3.24

RUN apk add --no-cache -U bash jq && \
    mkdir -p /etc/semaphore /var/lib/semaphore /tmp/semaphore

COPY --from=builder /go/src/semaphore/bin/semaphore /usr/local/bin/semaphore
COPY --from=builder /go/src/semaphore/deployment/docker/server/server-wrapper /usr/local/bin/server-wrapper
RUN chmod +x /usr/local/bin/semaphore /usr/local/bin/server-wrapper

EXPOSE 3000
CMD ["/usr/local/bin/server-wrapper"]
