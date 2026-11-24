# Docker

Generally we are building production-grade images for each tag, latest and even
for the development branch which will be pushed to [DockerHub][dockerhub]. If
you still need to build your own image you can easily do that, you just need
install [Docker][docker] and [Task][gotask] on your system.

If you just want to use our pre-built images please follow the instructions on
our [documentation][documentation].

If you want to use [docker-compose][dockercompose] to start Semaphore you could
also read about it on our [documentation][documentation] or take a look at our
collection of [snippets][snippets] within this repository.

## Build

We have prepared multiple tasks to build an publish container images, including
tasks to verify the image contains all required tools:

```console
task docker:build
task docker:push
```

If you want to customize the image names or if you want to use [Podman][podman]
instead of [Docker][docker] you are able to provide some set of environment
variables to the [Task][gotask] command:

* `DOCKER_ORG`: Define a custom organization for the image, defaults to `semaphoreui`
* `DOCKER_SERVER`: Define a different name for the server image, defaults to `semaphore`
* `DOCKER_RUNNER`: Define a different name for the runner image, defaults to `runner`
* `DOCKER_CMD`: Use another command to build the image, defaults to `docker`

## Customizing Version Information

If you are building downstream Docker images based on Semaphore and want to display your own version string in the Semaphore UI, you can use the `SEMAPHORE_BUILD_INFO` environment variable:

```console
# Set custom version info at runtime
docker run -e SEMAPHORE_BUILD_INFO="v2.14.9-mycompany-1.0.0" semaphoreui/semaphore:v2.14.9

# Or set it when building your own image
FROM semaphoreui/semaphore:v2.14.9
ENV SEMAPHORE_BUILD_INFO="v2.14.9-downstream-custom"
```

When `SEMAPHORE_BUILD_INFO` is set, it will override the default version display format (`{version}-{commit}-{date}`) in both the CLI and web UI. This is useful for downstream distributions that want to show their own version information to users.

## Test

We defined tasks to handle some linting and to verify the images contain the
tools and binaries that are required to run Semaphore. Here we are using
[Hadolint][hadolint] to ensure we are mostly following best-practices and
[Goss][goss] which is using a configuration file to define the requirements.

To install the required tools you also need to install [Golang][golang] on your
system, the installation of [Golang][golang] is not covered by us.

The installation of the dependencies can be customized by providing environment
variables for `INSTALL_PATH` (`/usr/local/bin`) and `REQUIRE_SUDO` (true).

```console
task docker:test
```

[dockerhub]: https://hub.docker.com/r/semaphoreui/semaphore
[docker]: https://docs.docker.com/engine/install/
[podman]: https://podman.io/docs/installation
[gotask]: https://taskfile.dev/installation/
[dockercompose]: https://docs.docker.com/compose/
[golang]: https://go.dev/doc/install
[hadolint]: https://github.com/hadolint/hadolint
[goss]: https://github.com/goss-org/goss
[snippets]: ../compose/README.md
[documentation]: https://docs.semaphoreui.com/administration-guide/installation
