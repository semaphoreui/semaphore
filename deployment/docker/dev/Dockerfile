# Golang testing image with some tools already installed
FROM golang:1.13-alpine

ENV SEMAPHORE_VERSION="development" SEMAPHORE_ARCH="linux_amd64" \
    SEMAPHORE_CONFIG_PATH="${SEMAPHORE_CONFIG_PATH:-/etc/semaphore}" \
    APP_ROOT="/go/src/github.com/ansible-semaphore/semaphore/"

RUN apk add --no-cache git mysql-client python py-pip py-openssl openssl ca-certificates curl curl-dev openssh-client tini nodejs nodejs-npm bash rsync && \
    apk --update add --virtual build-dependencies python-dev libffi-dev openssl-dev build-base &&\
    rm -rf /var/cache/apk/*

RUN pip install --upgrade pip cffi   && \
    pip install ansible &&\
    apk del build-dependencies

RUN adduser -D -u 1002 -g 0 semaphore && \
    mkdir -p /go/src/github.com/ansible-semaphore/semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    chown -R semaphore:0 /go && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore && \
    ssh-keygen -t rsa -q -f "/root/.ssh/id_rsa" -N ""       && \
    ssh-keyscan -H github.com > /root/.ssh/known_hosts

RUN cd $(go env GOPATH) && curl -sL https://taskfile.dev/install.sh | sh

# Copy in app source
WORKDIR ${APP_ROOT}
COPY . ${APP_ROOT}
RUN deployment/docker/dev/bin/install

USER 1000
EXPOSE 3000

CMD ["task", "watch"]
