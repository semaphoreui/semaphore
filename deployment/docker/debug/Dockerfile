FROM --platform=$BUILDPLATFORM golang:1.22-alpine3.18 as builder

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN apk add --no-cache -U \
    libc-dev curl nodejs npm git gcc zip unzip tar

WORKDIR /usr/local
# hadolint ignore=DL4006
RUN curl -sL https://taskfile.dev/install.sh | sh

WORKDIR /go/src/semaphore
COPY go.mod go.sum /go/src/semaphore/

RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

COPY . /go/src/semaphore

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task deps && \
    task build:debug GOOS=${TARGETOS} GOARCH=${TARGETARCH}

FROM alpine:3.19

RUN apk add --no-cache -U \
bash curl git gnupg mysql-client openssh-client-default python3 python3-dev py3-pip rsync sshpass tar tini tzdata unzip wget zip build-base openssl-dev libffi-dev cargo && \
    rm -rf /var/cache/apk/* && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    mkdir -p /var/lib/semaphore && \
    mkdir -p /opt/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore && \
    chown -R semaphore:0 /var/lib/semaphore && \
    chown -R semaphore:0 /opt/semaphore

COPY --chown=1001:0 ./deployment/docker/debug/ansible.cfg /etc/ansible/ansible.cfg
COPY --from=builder /go/src/semaphore/deployment/docker/debug/server-wrapper /usr/local/bin/
COPY --from=builder /go/src/semaphore/bin/semaphore /usr/local/bin/
COPY --from=builder /go/bin/dlv /

RUN chown -R semaphore:0 /usr/local/bin/server-wrapper && \
    chmod +x /usr/local/bin/server-wrapper && \
    chown -R semaphore:0 /usr/local/bin/semaphore && \
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore
USER 1001

# renovate: datasource=pypi depName=ansible
ENV ANSIBLE_VERSION 9.4.0

RUN mkdir /opt/semaphore/venv

RUN python3 -m venv /opt/semaphore/venv --system-site-packages && \
    source /opt/semaphore/venv/bin/activate && \
    pip3 install --upgrade pip ansible==${ANSIBLE_VERSION} boto3 botocore requests

# Preventing ansible zombie processes. Tini kills zombies.
ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/usr/local/bin/server-wrapper"]