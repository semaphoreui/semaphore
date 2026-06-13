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


RUN pip3 install --no-cache-dir --break-system-packages \
        "ansible-core>=2.16,<2.18" \
        "ansible==${ANSIBLE_VERSION}" \
        jmespath \
        netaddr \
        passlib

RUN set -eux; \
    for bin in sh cat chmod rm ls echo ssh ssh-agent ssh-add setsid git ansible-playbook python3; do \
        command -v "$bin" >/dev/null || { echo "missing: $bin"; exit 1; }; \
    done; \
    setsid -w true

RUN install -d -o nobody -g 0 -m 0775 /workspace \
 && chmod g=u /etc/passwd

WORKDIR /workspace
USER nobody:0

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["sh"]