# ansible-semaphore production image
#
# Uses frolvlad alpine so we have access to glibc which is needed for golang
# and when deploying in openshift
FROM frolvlad/alpine-glibc:alpine-3.7

LABEL maintainer="Tom Whiston <tom.whiston@gmail.com>"

RUN apk add --no-cache git ansible mysql-client curl openssh-client tini && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore

COPY ./ /go/src/github.com/ansible-semaphore/semaphore
WORKDIR /go/src/github.com/ansible-semaphore/semaphore

RUN apk add --no-cache -U libc-dev go nodejs && \
  ./deployment/docker/prod/bin/install && \
  apk del libc-dev go nodejs && \
  rm -rf /go/* && \
  rm -rf /var/cache/apk/*

WORKDIR /home/semaphore
USER 1001

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/semaphore-wrapper", "/usr/local/bin/semaphore", "--config", "/etc/semaphore/config.json"]
