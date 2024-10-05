# Semaphore UI

Modern UI for Ansible, Terraform, OpenTofu, PowerShell and other DevOps tools.

[![telegram](https://img.shields.io/badge/discord_community-skyblue?style=for-the-badge&logo=discord)](https://discord.gg/5R6k7hNGcH) 
[![telegram](https://img.shields.io/badge/youtube_channel-red?style=for-the-badge&logo=youtube)](https://www.youtube.com/@semaphoreui) 
<!-- [![docker](https://img.shields.io/badge/container_configurator-white?style=for-the-badge&logo=docker)](https://semaphoreui.com/install/docker/) -->

![responsive-ui-phone1](https://user-images.githubusercontent.com/914224/134777345-8789d9e4-ff0d-439c-b80e-ddc56b74fcee.png)

If your project has grown and deploying from the terminal is no longer feasible, then Semaphore UI is the tool you need.

## Live Demo

Try the latest version of Semaphore at [https://cloud.semaphoreui.com](https://cloud.semaphoreui.com).


## What is Semaphore UI?

Semaphore UI is a modern web interface for managing popular DevOps tools.

Semaphore UI allows you to:
* Easily run Ansible playbooks, Terraform and OpenTofu code, as well as Bash and PowerShell scripts.
* Receive notifications about failed tasks.
* Control access to your deployment system.

## Key Concepts

1. **Projects** is a collection of related resources, configurations, and tasks. Each project allows you to organize and manage your automation efforts in one place, defining the scope of tasks such as deploying applications, running scripts, or orchestrating cloud resources. Projects help group resources, inventories, task templates, and environments for streamlined automation workflows.
2. **Task Templates** are reusable definitions of tasks that can be executed on demand or scheduled. A template specifies what actions should be performed, such as running Ansible playbooks, Terraform configurations, or other automation tasks. By using templates, you can standardize tasks and easily re-execute them with minimal effort, ensuring consistent results across different environments.
3. **Task** is a specific instance of a job or operation executed by Semaphore. It refers to running a predefined action (like an Ansible playbook or a script) using a task template. Tasks can be initiated manually or automatically through schedules and are tracked to give you detailed feedback on the execution, including success, failure, and logs.
4. **Schedules** allow you to automate task execution at specified times or intervals. This feature is useful for running periodic maintenance tasks, backups, or deployments without manual intervention. You can configure recurring schedules to ensure important automation tasks are performed regularly and on time.
5. **Inventory** is a collection of target hosts (servers, virtual machines, containers, etc.) on which tasks will be executed. The inventory includes details about the managed nodes such as IP addresses, SSH credentials, and grouping information. It allows for dynamic control over which environments and hosts your automation will interact with.
6. **Environment** refers to a configuration context that holds sensitive information such as environment variables and secrets used by tasks during execution. It separates sensitive data from task templates and allows you to switch between different setups while running the same task template across different environments securely.

## Getting Started

You can install Semaphore using the following methods:
* Docker
* SaaS ([Semaphore Cloud](https://cloud.semaphoreui.com))
* Deploy a VM from a marketplace (AWS, DigitalOcean, etc.)
* Snap
* Binary file
* Debian or RPM package

### Docker

The most popular way to install Semaphore is via Docker.

```
docker run -p 3000:3000 --name semaphore \
	-e SEMAPHORE_DB_DIALECT=bolt \
	-e SEMAPHORE_ADMIN=admin \
	-e SEMAPHORE_ADMIN_PASSWORD=changeme \
	-e SEMAPHORE_ADMIN_NAME=Admin \
	-e SEMAPHORE_ADMIN_EMAIL=admin@localhost \
	-d semaphoreui/semaphore:latest
```

We recommend using the [Container Configurator](https://semaphoreui.com/install/docker/) to get the ideal Docker configuration for Semaphore.

### SaaS

We offer a SaaS solution for using Semaphore UI without installation. Check it out at [Semaphore Cloud](https://cloud.semaphoreui.com).

### Deploy VM from Marketplace

Supported cloud providers:
* [Semaphore Run](https://cloud.semaphore.run/servers/new/semaphore)
* [AWS](https://aws.amazon.com/marketplace/pp/prodview-5noeat2jipwca)
* [Yandex Cloud](https://yandex.cloud/en-ru/marketplace/products/fastlix/semaphore)
* DigitalOcean (coming soon)

### Other Installation Methods

For more installation options, visit our [Installation page](https://semaphoreui.com/install).

## Documentation

* [User Guide](https://docs.semaphoreui.com)
* [API Reference](https://semaphoreui.com/api-docs)

## License
MIT © [Denis Gukov](https://github.com/fiftin)

[![patreon](https://img.shields.io/badge/become_a_patreon-teal?style=for-the-badge&logo=patreon)](https://www.patreon.com/semaphoreui) 
[![ko-fi](https://img.shields.io/badge/buy_me_a_coffee-pink?style=for-the-badge&logo=kofi)](https://ko-fi.com/fiftin) 
