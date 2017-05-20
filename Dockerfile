FROM alpine:3.5

ENV SEMAPHORE_VERSION="2.3.0" SEMAPHORE_ARCH="linux_amd64"

RUN apk add --no-cache git ansible mysql-client curl openssh-client tini && \
    curl -sSfL "https://github.com/ansible-semaphore/semaphore/releases/download/v$SEMAPHORE_VERSION/semaphore_$SEMAPHORE_ARCH" > /usr/bin/semaphore && \
    chmod +x /usr/bin/semaphore && mkdir -p /etc/semaphore/playbooks

EXPOSE 3000

ADD ./scripts/docker-startup.sh /usr/bin/semaphore-startup.sh

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["/usr/bin/semaphore-startup.sh", "/usr/bin/semaphore", "-config", "/etc/semaphore/semaphore_config.json"]
