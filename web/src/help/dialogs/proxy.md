---
form: src/components/ProxyForm.vue
---

# Edit Proxy

A proxy is a host that the tasks of the project pass through to reach servers they can not
connect to directly. Choose it in the **Proxy** field of a repository, to fetch the repository
through it, or of an inventory, to connect Ansible to the hosts through it.

## Name {#proxyForm-name}

Required, and unique in the project. Repositories and inventories show the proxy by this name.

## Type {#proxyForm-type}

- **SSH (jump host)**: Semaphore connects to the host over SSH and opens the connection to the
  server from there. Only this type can require another proxy.
- **SOCKS5**, **HTTP**, **HTTPS**: a proxy server that carries the connection. Git uses it itself
  for an `https` repository; Semaphore passes an SSH connection through it, so no other tool has
  to be installed.

## Host {#proxyForm-host}

Required. A host name or an IP address, without a user, a port or a scheme. A host name may
contain only letters, digits, dots, hyphens and underscores.

## Port {#proxyForm-port}

Optional, from 1 to 65535. Empty means `22` for an SSH jump host. Give the port for the other
types: without one, Git and SSH connections assume different ports.

## User {#proxyForm-user}

SSH only: the login on the jump host. Empty means the user Semaphore runs as. Letters, digits,
dots, hyphens and underscores, not starting with a hyphen. A SOCKS5 or HTTP proxy takes its login
from **Access Key** instead.

## Access Key {#proxyForm-sshKeyId}

Optional. The key from the Key Store that the proxy itself is logged in to with, usually not the
key of the servers behind it:

- an **SSH Key** for an SSH jump host;
- a **Login with password** for SOCKS5, HTTP and HTTPS. The password is passed in the environment,
  never on a command line.

## Requires Proxy {#proxyForm-requiresProxyId}

SSH only, optional. Another SSH proxy of the project to pass through first, for a jump host that
is reachable only from another one. A chain holds at most 5 proxies and can not loop back to a
proxy already in it.
