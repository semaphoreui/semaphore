FROM golang:1.22-alpine3.19 as golang

RUN apk add --no-cache -U \
    curl git

WORKDIR /usr/local
# hadolint ignore=DL4006
RUN curl -sL https://taskfile.dev/install.sh | sh

WORKDIR /go/src/semaphore
COPY go.mod go.sum /go/src/semaphore/

RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

COPY . /go/src/semaphore

RUN --mount=type=cache,target=/go/pkg --mount=type=cache,target=/root/.cache/go-build \
    task deps:tools && \
    task deps:be && \
    task e2e:goodman && \
    task e2e:hooks

FROM apiaryio/dredd:13.0.0

RUN apk add --no-cache -U \
    bash git go

COPY --from=golang /go/bin/goodman /root/go/bin/goodman
COPY --from=golang /go/src/semaphore /semaphore
WORKDIR /semaphore

COPY deployment/docker/dredd/entrypoint /usr/local/bin
ENTRYPOINT ["/usr/local/bin/entrypoint"]
