# ansible-semaphore production image
FROM --platform=$BUILDPLATFORM golang:1.21-alpine3.18 as builder

COPY ./ /go/src/github.com/ansible-semaphore/semaphore
WORKDIR /go/src/github.com/ansible-semaphore/semaphore

ARG TARGETOS
ARG TARGETARCH
ARG TERRAFORM_VERSION="1.8.2"

RUN apk add --no-cache -U libc-dev curl nodejs npm git gcc unzip
RUN ./deployment/docker/prod/bin/install ${TARGETOS} ${TARGETARCH}

RUN curl -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip
RUN unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin
RUN rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip

FROM alpine:3.18 as runner
LABEL maintainer="Tom Whiston <tom.whiston@gmail.com>"

RUN apk add --no-cache bash sshpass git curl ansible mysql-client openssh-client-default tini py3-aiohttp tzdata py3-pip && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    mkdir -p /var/lib/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore && \
    chown -R semaphore:0 /var/lib/semaphore

COPY --from=builder /usr/local/bin/semaphore-wrapper /usr/local/bin/
COPY --from=builder /usr/local/bin/semaphore /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/semaphore-wrapper &&\
    chown -R semaphore:0 /usr/local/bin/semaphore &&\
    chmod +x /usr/local/bin/semaphore-wrapper &&\
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore
USER 1001

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/semaphore-wrapper", "/usr/local/bin/semaphore", "server", "--config", "/etc/semaphore/config.json"]
