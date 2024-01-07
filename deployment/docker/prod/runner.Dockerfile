FROM dind-ansible:latest

RUN apk add --no-cache wget git rsync

RUN adduser -D -u 1001 -G root -G docker semaphore && \
    mkdir -p /tmp/semaphore && \
    mkdir -p /etc/semaphore && \
    mkdir -p /var/lib/semaphore && \
    chown -R semaphore:0 /tmp/semaphore && \
    chown -R semaphore:0 /etc/semaphore && \
    chown -R semaphore:0 /var/lib/semaphore

RUN wget https://raw.githubusercontent.com/ansible-semaphore/semaphore/develop/deployment/docker/common/runner-wrapper -P /usr/local/bin/ && chmod +x /usr/local/bin/runner-wrapper
RUN wget https://github.com/ansible-semaphore/semaphore/releases/download/v2.9.37/semaphore_2.9.37_linux_amd64.tar.gz -O - | tar -xz -C /usr/local/bin/ semaphore

RUN chown -R semaphore:0 /usr/local/bin/runner-wrapper &&\
    chown -R semaphore:0 /usr/local/bin/semaphore &&\
    chmod +x /usr/local/bin/runner-wrapper &&\
    chmod +x /usr/local/bin/semaphore

WORKDIR /home/semaphore
USER 1001

RUN mkdir ./venv

RUN python3 -m venv ./venv --system-site-packages && \
    source ./venv/bin/activate && \
    pip3 install --upgrade pip boto3 botocore requests

RUN echo '{"tmp_path": "/tmp/semaphore","dialect": "bolt", "runner": {"config_file": "/var/lib/semaphore/runner.json"}}' > /etc/semaphore/config.json

CMD [ "/usr/local/bin/runner-wrapper" ]