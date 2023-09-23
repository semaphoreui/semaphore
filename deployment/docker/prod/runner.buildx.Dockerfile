# ansible-semaphore production image
FROM --platform=$BUILDPLATFORM golang:1.19-alpine3.18 as builder

COPY ./ /go/src/github.com/ansible-semaphore/semaphore
WORKDIR /go/src/github.com/ansible-semaphore/semaphore

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache -U libc-dev curl nodejs npm git gcc
RUN ./deployment/docker/prod/bin/install ${TARGETOS} ${TARGETARCH}

FROM alpine/ansible:latest

RUN apk add --no-cache wget git rsync

RUN adduser -D -u 1001 -G root semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    mkdir -p /var/lib/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore && \
    chown -R semaphore:0 /var/lib/semaphore

COPY --from=builder /usr/local/bin/runner-wrapper /usr/local/bin/
COPY --from=builder /usr/local/bin/semaphore /usr/local/bin/

RUN chown -R semaphore:0 /usr/local/bin/runner-wrapper &&\
    chown -R semaphore:0 /usr/local/bin/semaphore &&\
    chmod +x /usr/local/bin/runner-wrapper &&\
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore
USER 1001

RUN mkdir ./venv

RUN python3 -m venv ./venv --system-site-packages && \
    source ./venv/bin/activate && \
    pip3 install --upgrade pip

RUN pip3 install boto3 botocore

RUN echo '{"tmp_path": "/tmp/semaphore","dialect": "bolt", "runner": {"config_file": "/var/lib/semaphore/runner.json"}}' > /etc/semaphore/config.json

CMD [ "/usr/local/bin/runner-wrapper" ]