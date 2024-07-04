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


ENV OPENTOFU_VERSION="1.7.0"
ENV TERRAFORM_VERSION="1.8.2"
#ENV PULUMI_VERSION="3.116.1"

RUN wget https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz && \
    tar xf tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz -C /tmp && \
    rm tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz

RUN curl -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip -d /tmp && \
    rm terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip

#RUN if [ "$TARGETARCH" = "amd64" ]; then \
#      export PULUMI_ARCH="x64"; \
#    else \
#      export PULUMI_ARCH="${TARGETARCH}"; \
#    fi && \
#    wget -O pulumi.tar.gz https://github.com/pulumi/pulumi/releases/download/v${PULUMI_VERSION}/pulumi-v${PULUMI_VERSION}-linux-${PULUMI_ARCH}.tar.gz
#RUN tar xf pulumi.tar.gz --strip-components=1 -C /usr/local/bin
#RUN rm pulumi.tar.gz

FROM alpine:3.19

ARG TARGETARCH="amd64"

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
    chown -R semaphore:0 /opt/semaphore && \
    find /usr/lib/python* -iname __pycache__ | xargs rm -rf

COPY --chown=1001:0 ./deployment/docker/runner/ansible.cfg /etc/ansible/ansible.cfg
COPY --from=builder /go/src/semaphore/deployment/docker/runner/runner-wrapper /usr/local/bin/
COPY --from=builder /go/src/semaphore/bin/semaphore /usr/local/bin/
COPY --from=builder /tmp/tofu /usr/local/bin/
COPY --from=builder /tmp/terraform /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/runner-wrapper && \
    chmod +x /usr/local/bin/runner-wrapper && \
    chown -R semaphore:0 /usr/local/bin/semaphore && \
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore

# renovate: datasource=pypi depName=ansible
ENV ANSIBLE_VERSION 9.4.0
ARG ANSIBLE_VENV_PATH=/opt/semaphore/apps/ansible/${ANSIBLE_VERSION}/venv

RUN apk add --no-cache -U python3-dev build-base openssl-dev libffi-dev cargo && \
     mkdir -p ${ANSIBLE_VENV_PATH} && \
     python3 -m venv ${ANSIBLE_VENV_PATH} --system-site-packages && \
     source ${ANSIBLE_VENV_PATH}/bin/activate && \
     pip3 install --upgrade pip ansible==${ANSIBLE_VERSION} boto3 botocore requests && \
     apk del python3-dev build-base openssl-dev libffi-dev cargo && \
     rm -rf /var/cache/apk/* && \
     find ${ANSIBLE_VENV_PATH} -iname __pycache__ | xargs rm -rf && \
     chown -R semaphore:0 /opt/semaphore

USER 1001

ENV VIRTUAL_ENV="$ANSIBLE_VENV_PATH"
ENV PATH="$ANSIBLE_VENV_PATH/bin:$PATH"

# Preventing ansible zombie processes. Tini kills zombies.
ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/usr/local/bin/runner-wrapper"]
