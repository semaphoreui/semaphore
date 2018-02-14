FROM alpine:3.5

ARG SEMAPHORE_VERSION
ENV SEMAPHORE_VERSION=${SEMAPHORE_VERSION} SEMAPHORE_ARCH="linux_amd64"

RUN apk add --no-cache git ansible mysql-client curl openssh-client tini && \
    curl -sSfL "https://github.com/ansible-semaphore/semaphore/releases/download/v$SEMAPHORE_VERSION/semaphore_$SEMAPHORE_ARCH" > /usr/bin/semaphore && \
    chmod +x /usr/bin/semaphore && mkdir -p /etc/semaphore/playbooks

EXPOSE 3000

ADD ./common/semaphore-startup.sh /usr/bin/semaphore-startup.sh
RUN chmod +x /usr/bin/semaphore-startup.sh

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["/usr/bin/semaphore-startup.sh", "/usr/bin/semaphore", "-config", "/etc/semaphore/semaphore_config.json"]
