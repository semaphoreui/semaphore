# V2 branch

[![Circle CI](https://circleci.com/gh/ansible-semaphore/semaphore.svg?style=svg&circle-token=3702872acf2bec629017fa7dd99fdbea56aef7df)](https://circleci.com/gh/ansible-semaphore/semaphore)

Beware WIP

## Requirements

- GIT installed and in $PATH
- Ansible installed and in $PATH
- Redis & MySQL/MariaDB

## Getting started

```
$ ./semaphore -printConfig > config.json
$ vim config.json
$ ./semaphore -migrate -config `path`/config.json
... migrations exec'd ...
$ ./semaphore -hash myPassword
... hash printed ...
$ add user with mysql
$ ./semaphore -config `path`/config.json
... listening on a port
```