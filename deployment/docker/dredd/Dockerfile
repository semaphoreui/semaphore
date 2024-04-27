FROM golang:1.22-alpine3.19 as golang

RUN apk add --no-cache -U \
    curl git

WORKDIR /usr/local
# hadolint ignore=DL4006
RUN curl -sL https://taskfile.dev/install.sh | sh

COPY . /semaphore/
WORKDIR /semaphore

RUN task deps:tools && \
    task deps:be && \
    task e2e:hooks

FROM apiaryio/dredd:13.0.0

RUN apk add --no-cache -U \
    bash git

COPY --from=golang /go/bin/goodman /usr/local/bin/goodman
COPY --from=golang /semaphore /semaphore
WORKDIR /semaphore

COPY deployment/docker/dredd/entrypoint /usr/local/bin
ENTRYPOINT ["/usr/local/bin/entrypoint"]
