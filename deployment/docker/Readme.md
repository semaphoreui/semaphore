# Docker Deployments

Production images for each tag, latest and the development branch will be pushed to [docker hub](https://hub.docker.com/r/castawaylabs/semaphore).
To build images locally see the contexts included here and use the `d` and `dc` tasks in the root Taskfile.yml to help with building and running.

## Contexts

### Prod

To build a production image you should run

    context=prod task docker:build
    
this will create an image called `castawaylabs/semaphore:latest` which will be compiled from the currently checked out code

This image is run as non root user 1001 (for PaaS systems such as openshift) and is build on alpine with added glibc.
With ansible etc... installed in the container it is ~283MiB in size.

You will need to provide environmental variables so that the configuration can be built correctly for your environment.
See `docker-compose.yml` for an example, or look at `../common/entrypoint` to see which variables are available
        
If you want to bulid an image with a custom tag you can optionally pass a tag to the command

    context=prod tag=mybranch task docker:build
    
#### Example Configuration

To run Semaphore in a simple production-like docker configuration run the following command:

    task dc:prod

You can then access Semaphore directly from the url http://localhost:8081/

#### SSL Termination Using Nginx

Generate a cert, ca cert, and key file and place into `prod/proxy/cert/` with
these names:

* `cert.pem`
* `privkey.pem`
* `fullchain.pem`

(I've used letsencrypt generated certs with success.)

Run `task dc:prod` and your Semaphore instance will then be at the url
https://localhost:8443

If you do not add certificates the container will create self-signed certs instead

## Dev

To start a development start you could run
```
context=dev task dc:up
```
The development stack will run `task watch` by default and `dc:up` will volume link the application in to the container.
Without `dc:up` the application will run the version of the application which existed at image build time.

The development container is based on [micro-golang](https://github.com/twhiston/micro-golang)'s test base image
which contains the go toolchain and glibc in alpine.

Because the test image links your local volume it expects that you have run `task deps` and `task compile` locally 
as necessary to make the application usable.

## Convenience Functions

### dc:dev

`dc:dev` rebuilds the development images and runs a development stack, with the semaphore root as a volume link
This allows you to work inside the container with live code. The container has all the tools you need to build and test semaphore

### dc:prod
   
`dc:prod` rebuilds the production example images and starts the production-like stack. 
This will compile the application for the currently checked out code but will not leave build tools or source in the container.
Therefore file changes will result in needing a rebuild.