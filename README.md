# Semaphore UI (formerly Ansible Semaphore)

[![docker](https://img.shields.io/badge/docker_container_configurator-skyblue?style=for-the-badge&logo=docker)](https://semaphoreui.com/install/docker/)
[![patreon](https://img.shields.io/badge/support_semaphore-teal?style=for-the-badge&logo=patreon)](https://www.patreon.com/semaphoreui) 
[![telegram](https://img.shields.io/badge/telegram_community-blue?style=for-the-badge&logo=telegram)](https://t.me/semaphoreui) 
[![telegram](https://img.shields.io/badge/youtube_channel-red?style=for-the-badge&logo=youtube)](https://www.youtube.com/@semaphoreui) 

Semaphore is a modern UI for Ansible, Terraform/OpenTofu, Bash and Pulumi. It lets you easily run Ansible playbooks, get notifications about fails, control access to deployment system.

If your project has grown and deploying from the terminal is no longer for you then Semaphore UI is what you need.

![responsive-ui-phone1](https://user-images.githubusercontent.com/914224/134777345-8789d9e4-ff0d-439c-b80e-ddc56b74fcee.png)

## Installation

### Docker

https://hub.docker.com/r/semaphoreui/semaphore

`docker-compose.yml` for minimal configuration:

```yaml
services:
  semaphore:
    ports:
      - 3000:3000
    image: semaphoreui/semaphore:latest
    environment:
      SEMAPHORE_DB_DIALECT: bolt
      SEMAPHORE_ADMIN_PASSWORD: changeme
      SEMAPHORE_ADMIN_NAME: admin
      SEMAPHORE_ADMIN_EMAIL: admin@localhost
      SEMAPHORE_ADMIN: admin
      TZ: Europe/Berlin
    volumes:
      - /path/to/data/home:/etc/semaphore # config.json location
      - /path/to/data/lib:/var/lib/semaphore # database.boltdb location (Not required if using mysql or postgres)
```

### Other installation methods
https://docs.semaphoreui.com/administration-guide/installation

## Demo

You can test latest version of Semaphore on https://cloud.semui.co.

## Docs

Admin and user docs: https://docs.semaphoreui.com.

API description: https://semaphoreui.com/api-docs/.
