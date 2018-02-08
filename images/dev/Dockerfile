# Golang testing image with some tools already installed
FROM tomwhiston/micro-golang:test-base

LABEL maintainer="Tom Whiston <tom.whiston@gmail.com>"

ENV SEMAPHORE_VERSION="development" SEMAPHORE_ARCH="linux_amd64"
ENV WATCH_STARTUP_CMD="images/common/semaphore-startup.sh /tmp/semaphore -config /etc/semaphore/semaphore_config.json"

RUN apk add --no-cache git ansible mysql-client curl openssh-client tini nodejs nodejs-npm bash && \
    mkdir -p /etc/semaphore/playbooks && \
    mkdir -p /go/src/app /go/src/github.com/ansible-semaphore/semaphore && \
    ln -s /go/src/github.com/ansible-semaphore/semaphore /go/src/app && \
    ssh-keygen -t rsa -q -f "/root/.ssh/id_rsa" -N "" && \
    ssh-keyscan -H github.com > /root/.ssh/known_hosts

# Add a default config which can be overriden with a file in app-root
RUN echo '{\n\
        "mysql": {\n\
            "host": "mysql:3306",\n\
            "user": "semaphore",\n\
            "pass": "semaphore",\n\
            "name": "semaphore"\n\
        },\n\
        "port": ":3000"\n\
    }\n'\
    >> /go/src/app/config.json

EXPOSE 3000

# Copy in app source
COPY . /go/src/github.com/ansible-semaphore/semaphore
WORKDIR /go/src/github.com/ansible-semaphore/semaphore

CMD ["./make.sh", "watch"]