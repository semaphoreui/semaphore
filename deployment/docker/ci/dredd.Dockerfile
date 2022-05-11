FROM golang:1.18.2-alpine3.15 as golang

RUN apk add --no-cache curl git

# We need the source and task to compile the hooks
COPY . /semaphore/

RUN (cd /usr && curl -sL https://taskfile.dev/install.sh | sh)
WORKDIR /semaphore
RUN task deps:tools && task deps:be && task compile:be && task compile:api:hooks

FROM apiaryio/dredd:13.0.0 as dredd

RUN apk add --no-cache bash go git

RUN go get github.com/snikch/goodman/cmd/goodman

COPY --from=golang /semaphore /semaphore

WORKDIR /semaphore

COPY deployment/docker/ci/dredd/entrypoint /usr/local/bin

ENV SEMAPHORE_SERVICE=semaphore_ci \
    SEMAPHORE_PORT=3000 \
    MYSQL_SERVICE=mysql \
    MYSQL_PORT=3306

ENTRYPOINT ["/usr/local/bin/entrypoint"]
