FROM debian:13-slim

ARG ANSIBLE_VERSION=9.5.1
ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
        openssh-client \
        util-linux \
        git \
        ca-certificates \
        python3 \
        python3-pip \
        python3-venv \
        sshpass \
        tini \
        curl \
    && rm -rf /var/lib/apt/lists/*

RUN install -d -o nobody -g 0 -m 0775 /workspace \
 && chmod g=u /etc/passwd

WORKDIR /workspace
USER nobody:0

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["sh"]