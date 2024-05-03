FROM --platform=$BUILDPLATFORM golang:1.22-alpine3.19 as builder

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

RUN --mount=type=cache,target=/go/src/semaphore/web/node_modules \
    --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task deps && \
    task build GOOS=${TARGETOS} GOARCH=${TARGETARCH}

FROM alpine:3.19

RUN apk add --no-cache -U \
    bash curl git mysql-client openssh-client-default python3 py3-pip rsync sshpass tar tini tzdata unzip wget zip gcc && \
    rm -rf /var/cache/apk/* && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    mkdir -p /var/lib/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore && \
    chown -R semaphore:0 /var/lib/semaphore

COPY --chown=1001:0 ./deployment/docker/server/ansible.cfg /tmp/semaphore/ansible.cfg
COPY --from=builder /go/src/semaphore/deployment/docker/server/server-wrapper /usr/local/bin/
COPY --from=builder /go/src/semaphore/bin/semaphore /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/server-wrapper && \
    chmod +x /usr/local/bin/server-wrapper && \
    chown -R semaphore:0 /usr/local/bin/semaphore && \
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore
USER 1001

# renovate: datasource=pypi depName=ansible
ENV ANSIBLE_VERSION 9.4.0

ENV PATH /home/semaphore/.local/bin:$PATH

# hadolint ignore=DL3013
RUN pip3 install -U --break-system-packages --no-cache-dir \
    ansible==${ANSIBLE_VERSION} \
    boto3 botocore requests netaddr

ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/usr/local/bin/server-wrapper"]
