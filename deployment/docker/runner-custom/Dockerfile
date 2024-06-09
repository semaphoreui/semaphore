FROM --platform=$BUILDPLATFORM golang:1.22-alpine3.18 as builder

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
    task build GOOS=${TARGETOS} GOARCH=${TARGETARCH}

FROM alpine:3.19

RUN apk add --no-cache -U \
    bash curl git gnupg mysql-client openssh-client-default python3 python3-dev py3-pip rsync sshpass tar tini tzdata unzip wget zip && \
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

COPY --chown=1001:0 ./deployment/docker/runner/ansible.cfg /tmp/semaphore/ansible.cfg
COPY --from=builder /go/src/semaphore/deployment/docker/runner-custom/runner-wrapper /usr/local/bin/
COPY --from=builder /go/src/semaphore/bin/semaphore /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/runner-wrapper && \
    chmod +x /usr/local/bin/runner-wrapper && \
    chown -R semaphore:0 /usr/local/bin/semaphore && \
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore

# renovate: datasource=pypi depName=ansible
ENV ANSIBLE_VERSION 9.4.0

RUN apk add --no-cache -U build-base openssl-dev libffi-dev cargo && \
    mkdir /opt/semaphore/venv && \
    python3 -m venv /opt/semaphore/venv --system-site-packages && \
    source /opt/semaphore/venv/bin/activate && \
    pip3 install --upgrade pip ansible==${ANSIBLE_VERSION} boto3 botocore requests && \
    rm -rf /var/cache/apk/* && \
    find /opt/semaphore/venv -iname __pycache__ | xargs rm -rf && \
    chown -R semaphore:0 /opt/semaphore

ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/usr/local/bin/runner-wrapper"]
