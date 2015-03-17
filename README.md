semaphore
=========

Open Source Alternative to Ansible Tower

![screenshot](public/img/screenshot.png)

Features
--------

The basics of Ansible Tower, but in addition:

- [x] Fast, Simple interface (not having to submit a million forms to get something simple done)
- [x] Task output is streamed live via websocket
- [x] Create inventories per playbook
- [x] Add rsa keys (to authenticate git repositories)
- [x] Run playbooks against specified hosts
- [ ] Multiple Users support

Docker quickstart
-----------------

1. Get Docker
2. Run redis

```
docker run -d \
  --name=redisio \
  -v /var/lib/redisio:/var/lib/redis \
  -p 127.0.0.1:6379:6379 \
  castawaylabs/redis-docker
```

3. Run mongodb

```
docker run -d \
  --name=mongodb \
  -v /var/lib/mongodb:/var/lib/mongodb \
  -p 127.0.0.1:6379:6379 \
  castawaylabs/mongo-docker
```

4. Run semaphore

```
docker run -d \
  --name=semaphroe \
  --restart=always \
  --link redisio:redis \
  --link mongodb:mongo \
  -p 80:80 \
  castawaylabs/semaphore
```

Development steps:

Install requirements:
- node.js / io.js
- an isolated environment (e.g. Docker / Vagrant)
- ansible
- mongodb & redis

Initial Login
-------------

```
Email:			'admin@semaphore.local'
Password:		'CastawayLabs'
```

Environment Variables
---------------------

Use these variables to override the config.

| Variable Name | Description            | Default Value                   |
| ------------- | ---------------------- | ------------------------------- |
| PORT          | Web Port               | `80`                            |
| REDIS_PORT    | Redis Port             | `6379`                          |
| REDIS_HOST    | Redis Hostname         | `127.0.0.1`                     |
| REDIS_KEY     | Redis auth key         |                                 |
| BUGSNAG_KEY   | Bugsnag API key        |                                 |
| SMTP_USER     | Mandrill smtp username |                                 |
| SMTP_PASS     | Mandrill smtp password |                                 |
| MONGODB_URL   | Mongodb URL            | `mongodb://127.0.0.1/semaphore` |

Vision and goals
----------------

- Be able to specify environment information per playbook / per task
- Schedule jobs
- Email alerts
>>>>>>> 1199a52... Update site layout

Note to Ansible guys
--------------------

> Thanks very much for making Ansible, and Ansible Tower. It is a great tool!. Your UI is pretty horrible though, and so we'd be happy if you could learn and use parts of this tool in your Tower.

It would be amazing if this could be your `Community Edition` of Ansible Tower.

License
-------

MIT