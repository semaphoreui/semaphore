FROM debian:13-slim

ARG TARGETOS
ARG TARGETARCH
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
        wget \
        unzip \
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

ENV OPENTOFU_VERSION="1.11.0"
ENV TERRAFORM_VERSION="1.11.3"
ENV TERRAGRUNT_VERSION="0.78.0"

RUN wget https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz && \
    tar xf tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz -C /tmp && \
    rm tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.tar.gz

RUN curl -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip -d /tmp && \
    rm terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip

RUN wget -O /tmp/terragrunt https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_${TARGETARCH} && \
    chmod +x /tmp/terragrunt

RUN mv /tmp/tofu /usr/local/bin/
RUN mv /tmp/terraform /usr/local/bin/
RUN mv /tmp/terragrunt /usr/local/bin/

WORKDIR /workspace
USER nobody:0

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["sh"]