FROM apiaryio/dredd:13.0.0

ENV SEMAPHORE_SERVICE=semaphore_ci \
    SEMAPHORE_PORT=3000 \
    MYSQL_SERVICE=mysql \
    MYSQL_PORT=3306 \
    PATH=/root/go/bin:${PATH}

RUN apk add --no-cache bash curl git go

RUN (cd /usr && curl -sL https://taskfile.dev/install.sh | sh)

# We need the source and task to compile the hooks
COPY . /semaphore/

WORKDIR /semaphore

RUN task deps:tools
RUN task deps:be
RUN task compile:be
RUN task compile:api:hooks

COPY deployment/docker/ci/dredd/entrypoint /usr/local/bin

ENTRYPOINT ["/usr/local/bin/entrypoint"]
