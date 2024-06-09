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

ARG TARGETOS="linux"
ARG TARGETARCH="amd64"

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task deps && \
    task build GOOS=${TARGETOS} GOARCH=${TARGETARCH}

ENV OPENTOFU_VERSION="1.7.0"
ENV TERRAFORM_VERSION="1.8.2"
ENV PULUMI_VERSION="3.116.1"

RUN curl -sSLo tofu.tar.gz https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz && \
    tar xf tofu.tar.gz -C /tmp && \
    rm -f tofu.tar.gz

    # && \
    # curl -sSLo terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip && \
    # unzip terraform.zip -d /tmp && \
    # rm terraform.zip

FROM alpine:3.20

RUN apk add --no-cache -U \
    bash curl git gnupg mysql-client openssh-client-default rsync sshpass tar tini tzdata unzip wget zip python3 python3-dev py3-pip && \
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

ENV VIRTUAL_ENV=/opt/semaphore/venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

WORKDIR /home/semaphore

COPY --chown=1001:0 ./deployment/docker/runner/ansible.cfg /tmp/semaphore/ansible.cfg
COPY --from=builder /go/src/semaphore/deployment/docker/runner/runner-wrapper /usr/local/bin/
COPY --from=builder /go/src/semaphore/bin/semaphore /usr/local/bin/
COPY --from=builder /tmp/tofu /usr/local/bin/
# COPY --from=builder /tmp/terraform /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/runner-wrapper && \
    chmod +x /usr/local/bin/runner-wrapper && \
    chown -R semaphore:0 /usr/local/bin/semaphore && \
    chmod +x /usr/local/bin/semaphore

# renovate: datasource=pypi depName=ansible
ENV ANSIBLE_VERSION 9.4.0

RUN apk add --no-cache -U build-base openssl-dev libffi-dev cargo && \
    mkdir -p ${VIRTUAL_ENV} && \
    python3 -m venv ${VIRTUAL_ENV} --system-site-packages && \
    source ${VIRTUAL_ENV}/bin/activate && \
    pip3 install --upgrade pip ansible==${ANSIBLE_VERSION} boto3 botocore requests && \
    rm -rf /var/cache/apk/* && \
    find ${VIRTUAL_ENV} -iname __pycache__ | xargs rm -rf && \
    chown -R semaphore:0 /opt/semaphore

ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/usr/local/bin/runner-wrapper"]
