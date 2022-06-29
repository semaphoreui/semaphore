# Deploying Semaphore on Openshift

This is intended as a quick starter config to get semaphore up and running using only the docker hub image.The image is set to be periodically pulled from the repository source.

## Setup

Your openshift cluster needs to have the mysql-persistent template installed, however it comes by default.
```
# oc cluster up
oc new-project semaphore
oc create -fdeployment/openshift/template.yml
oc new-app mysql-persistent -p MYSQL_DATABASE=semaphore
oc new-app semaphore # -p SEMAPHORE_IMAGE_TAG=develop
```

It will take some moments for the application to become available (mainly due to the mysql pod startup time), check the logs of the semaphore container to see when it is ready. After this the web ui will be available on http://semaphore-semaphore.127.0.0.1.nip.io/auth/login (if running your oc cluster locally and you did not override the url via parameters). You can log in with the default values.
If you deploy the template to multiple namespaces you must set the SEMAPHORE_URL to a unique value or it will be rejected by the router.

## Parameters

`oc process --parameters semaphore`

|NAME| DESCRIPTION| VALUE|
|SEMAPHORE_IMAGE_SOURCE| The id of the repository from which to pull the semaphore image| docker.io/semaphoreui/semaphore|
|SEMAPHORE_IMAGE_TAG| The tag to use for the semaphore repository| latest|
|SEMAPHORE_DATA_VOLUME_SIZE| The size, in Gi of the semaphore data volume, which is mounted at /etc/semaphore| 5|
|SEMAPHORE_URL| Set this to the value which you wish to be passed to the route. Default value works for local development usage| semaphore-semaphore.127.0.0.1.nip.io|
