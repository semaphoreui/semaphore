FROM golang:1.13-alpine

ENV SEMAPHORE_VERSION="development" SEMAPHORE_ARCH="linux_amd64" \
    SEMAPHORE_CONFIG_PATH="${SEMAPHORE_CONFIG_PATH:-/etc/semaphore}" \
    APP_ROOT="/go/src/github.com/ansible-semaphore/semaphore/"

# hadolint ignore=DL3013
RUN apk add --no-cache git mysql-client python3 py3-pip py-openssl openssl ca-certificates curl curl-dev openssh-client tini nodejs nodejs-npm bash rsync && \
    apk --update add --virtual build-dependencies python3-dev libffi-dev openssl-dev build-base &&\
    rm -rf /var/cache/apk/*

RUN pip3 install --upgrade pip cffi   && \
    pip3 install ansible &&\
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

USER semaphore
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/semaphore-wrapper"]
CMD ["./bin/semaphore", "--config", "/etc/semaphore/config.json"]
