ARG SEMAPHORE_IMAGE
ARG SEMAPHORE_VERSION

FROM ${SEMAPHORE_IMAGE}:${SEMAPHORE_VERSION}

ARG TARGETARCH
ARG POWERSHELL_VERSION="7.5.0"

USER root

# Install dependencies
RUN apk add --no-cache \
    ca-certificates \
    less \
    ncurses-terminfo-base \
    krb5-libs \
    libgcc \
    libintl \
    libssl3 \
    libstdc++ \
    tzdata \
    userspace-rcu \
    zlib \
    icu-libs \
    curl

RUN apk -X https://dl-cdn.alpinelinux.org/alpine/edge/main add --no-cache \
    lttng-ust \
    openssh-client

RUN wget -O /tmp/powershell.tar.gz https://github.com/PowerShell/PowerShell/releases/download/v${POWERSHELL_VERSION}/powershell-${POWERSHELL_VERSION}-linux-musl-${TARGETARCH/amd/x}.tar.gz

RUN mkdir -p /opt/microsoft/powershell/${POWERSHELL_VERSION} \
    && tar zxf /tmp/powershell.tar.gz -C /opt/microsoft/powershell/${POWERSHELL_VERSION} \
    && rm /tmp/powershell.tar.gz \
    && chmod +x /opt/microsoft/powershell/${POWERSHELL_VERSION}/pwsh \
    && ln -s /opt/microsoft/powershell/${POWERSHELL_VERSION}/pwsh /usr/local/bin/pwsh \
    && ln -s /opt/microsoft/powershell/${POWERSHELL_VERSION}/pwsh /usr/local/bin/powershell

USER 1001