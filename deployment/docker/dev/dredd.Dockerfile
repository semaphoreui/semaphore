# hadolint ignore=DL3006
FROM tomwhiston/dredd:latest

ENV TASK_VERSION=v2.0.1 \
    GOPATH=/home/developer/go \
    SEMAPHORE_SERVICE=semaphore_dev \
    SEMAPHORE_PORT=3000 \
    MYSQL_SERVICE=mysql \
    MYSQL_PORT=3306

# We need the source and task to compile the hooks
USER 0
RUN dnf install -y nc
COPY deployment/docker/ci/dredd/entrypoint /usr/local/bin
COPY . /home/developer/go/src/github.com/ansible-semaphore/semaphore
WORKDIR /usr/local/bin
RUN curl -L "https://github.com/go-task/task/releases/download/${TASK_VERSION}/task_linux_amd64.tar.gz" | tar xvz && \
    chown -R developer /home/developer/go

# Get tools and do compile
WORKDIR /home/developer/go/src/github.com/ansible-semaphore/semaphore
RUN task deps:tools && task deps:be && task compile:be && task compile:api:hooks

ENTRYPOINT ["/usr/local/bin/entrypoint"]