# Golang testing image with some tools already installed
FROM tomwhiston/micro-golang:test-base

LABEL maintainer="Tom Whiston <tom.whiston@gmail.com>"

ENV SEMAPHORE_VERSION="development" SEMAPHORE_ARCH="linux_amd64" \
    WATCH_STARTUP_CMD="images/common/semaphore-startup.sh /tmp/semaphore -config /etc/semaphore/semaphore_config.json" \
    APP_ROOT="/go/src/github.com/ansible-semaphore/semaphore/"

RUN apk add --no-cache git mysql-client python py-pip openssl ca-certificates curl openssh-client tini nodejs nodejs-npm bash rsync && \
    apk --update add --virtual build-dependencies python-dev libffi-dev openssl-dev build-base  && \
    pip install --upgrade pip cffi    && \
    pip install ansible               && \
    mkdir -p /etc/semaphore/playbooks && \
    mkdir -p /go/src/app /go/src/github.com/ansible-semaphore/semaphore && \
    ln -s /go/src/github.com/ansible-semaphore/semaphore /go/src/app && \
    ssh-keygen -t rsa -q -f "/root/.ssh/id_rsa" -N "" && \
    ssh-keyscan -H github.com > /root/.ssh/known_hosts && \
    apk del build-dependencies            && \
    rm -rf /var/cache/apk/*


EXPOSE 3000

# Copy in app source
WORKDIR ${APP_ROOT}
COPY . .
RUN images/dev/bin/semaphore-bootstrap

CMD ["./make.sh", "watch"]