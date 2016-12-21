FROM alpine

RUN apk add --no-cache git ansible mysql-client curl openssh-client

RUN curl -L https://github.com/ansible-semaphore/semaphore/releases/download/v2.1.0/semaphore_linux_amd64 > /usr/bin/semaphore && chmod +x /usr/bin/semaphore && mkdir -p /etc/semaphore/playbooks

ADD semaphore-startup.sh /usr/bin/semaphore-startup.sh

RUN chmod +x /usr/bin/semaphore-startup.sh

EXPOSE 3000

ENTRYPOINT ["/usr/bin/semaphore-startup.sh"]

CMD ["/usr/bin/semaphore", "-config", "/etc/semaphore/semaphore_config.json"]
