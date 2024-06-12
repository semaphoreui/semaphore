# Semaphore UI


[![](https://img.shields.io/badge/Telegram-2CA5E0?style=flat-squeare&logo=telegram&logoColor=white)](https://t.me/semaphoreui)
[![](https://img.shields.io/youtube/channel/views/UCUjzgHjyeiiKsINaM6mHVQQ)](https://www.youtube.com/@semaphoreui)


[//]: # (![Website]&#40;https://img.shields.io/website?url=https%3A%2F%2Fsemui.co&#41;)


[//]: # ([![Twitter]&#40;https://img.shields.io/twitter/follow/semaphoreui?style=social&logo=twitter&#41;]&#40;https://twitter.com/semaphoreui&#41;)

[//]: # ([![ko-fi]&#40;https://ko-fi.com/img/githubbutton_sm.svg&#41;]&#40;https://ko-fi.com/fiftin&#41;)

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
https://docs.semui.co/administration-guide/installation

## Demo

You can test latest version of Semaphore on https://cloud.semui.co.

## Docs

Admin and user docs: https://docs.semui.co.

API description: https://semui.co/api-docs/.