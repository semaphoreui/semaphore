# ansible-semaphore production image
FROM golang:1.16.3-alpine3.13 as builder

LABEL maintainer="Tom Whiston <tom.whiston@gmail.com>"

COPY ./ /go/src/github.com/neo1908/semaphore
WORKDIR /go/src/github.com/neo1908/semaphore

RUN apk add --no-cache -U libc-dev curl nodejs npm git && \
  ./deployment/docker/prod/bin/install

# Uses frolvlad alpine so we have access to glibc which is needed for golang
# and when deploying in openshift
FROM frolvlad/alpine-glibc:alpine-3.13 as runner

RUN apk add --no-cache sshpass git curl ansible mysql-client openssh-client tini && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore

COPY --from=builder /usr/local/bin/semaphore-wrapper /usr/local/bin/
COPY --from=builder /usr/local/bin/semaphore /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/semaphore-wrapper &&\
    chown -R semaphore:0 /usr/local/bin/semaphore

WORKDIR /home/semaphore
USER 1001

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/semaphore-wrapper", "/usr/local/bin/semaphore", "--config", "/etc/semaphore/config.json"]
