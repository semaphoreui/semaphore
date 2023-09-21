# ansible-semaphore production image
FROM golang:1.19-alpine3.18 as builder

COPY ./ /go/src/github.com/ansible-semaphore/semaphore
WORKDIR /go/src/github.com/ansible-semaphore/semaphore

RUN apk add --no-cache -U libc-dev curl nodejs npm git gcc g++ && \
  ./deployment/docker/prod/bin/install

FROM alpine:3.18 as runner
LABEL maintainer="Denis Gukov <denguk@gmail.com>"

RUN apk add --no-cache sshpass git curl mysql-client openssh-client-default tini python3 py3-pip && \
    adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    mkdir -p /var/lib/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore && \
    chown -R semaphore:0 /var/lib/semaphore

# system dependencies, common jinja2-filters, common community-modules
RUN pip3 install ansible aiohttp \
    jmespath netaddr passlib \
    pywinrm requests cryptography && \
    touch /etc/semaphore/requirements.txt

COPY --from=builder /usr/local/bin/runner-wrapper /usr/local/bin/
COPY --from=builder /usr/local/bin/semaphore /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/runner-wrapper && \
    chown -R semaphore:0 /usr/local/bin/semaphore

WORKDIR /home/semaphore
USER 1001

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/runner-wrapper", "/usr/local/bin/semaphore", "runner", "--config", "/etc/semaphore/config.json"]
