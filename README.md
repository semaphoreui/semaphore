# V2 branch

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