# Dredd plus the Go hooks of .dredd/hooks, for the API contract tests
# (task dredd:docker). The hooks are compiled from the working tree.
FROM golang:1.26-alpine3.24 AS builder

RUN apk add --no-cache -U git

ENV GOWORK=off
WORKDIR /src

# go.mod replaces the pro stub with ./pro, so its module files must be
# present for the download to resolve.
COPY go.mod go.sum ./
COPY pro/go.mod pro/go.sum ./pro/
RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

# goodman is the Dredd hooks handler for Go; same source as `task dredd:goodman`.
RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/snikch/goodman/cmd/goodman@latest

COPY . .

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    cd .dredd/hooks && go build -o ../compiled_hooks

FROM apiaryio/dredd:14.0.0

RUN apk add --no-cache -U bash

# Dredd starts the Go hooks handler as $GOBIN/goodman.
ENV GOBIN=/usr/local/bin
COPY --from=builder /go/bin/goodman /usr/local/bin/goodman

WORKDIR /semaphore
COPY --from=builder /src/api-docs.yml ./api-docs.yml
COPY --from=builder /src/.dredd ./.dredd

COPY deployment/docker/dredd/entrypoint /usr/local/bin/entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint"]
