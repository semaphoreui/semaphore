# Deploying Semaphore on Openshift

This is intended as a quick starter config to get semaphore up and running using only the docker hub image.
The config is set to periodically pull new versions of the image from docker hub.

You may wish to tweak this once imported or choose to build the image within openshift itself

Your openshift cluster needs to have the mysql-persistent template installed, but it comes as default so it should be
```
oc new-project semaphore
oc create -fopenshift/template.yml
oc new-app mysql-persistent -p MYSQL_DATABASE=semaphore
oc new-app semaphore
```
It will take some moments for the application to become available (mainly due to the mysql pod startup time), after this the web ui will be available on http://semaphore-semaphore.127.0.0.1.nip.io/auth/login (if running your oc cluster locally and you did not override the url via parameters). You can login with the default values