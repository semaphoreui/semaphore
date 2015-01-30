semaphore
=========

Open Source Alternative to Ansible Tower

![](public/img/screenshot.png)

Features
--------

The basics of Ansible Tower, but in addition:

- Fast, Simple interface that doesn’t get in the way
- Task output is streamed live via websocket
- Free. MIT Licensed. Do what you want.

How to run:
-----------

1. Install Vagrant
2. Run `vagrant up`
3. Open [localhost:3000](http://localhost:3000)

Development steps:

Install requirements:
- node.js >= 0.11.x
- an isolated environment (e.g. Docker / NodeGear)
- ansible (the tool)
- mongodb & redis
- Sudo access (this might change). To run jobs, this tool writes private keys to /root/.ssh and copies playbook directories to /root/.

1. `cp lib/credentials.example.json lib/credentials.json` (<- make custom to your environment)
2. `bower install`
3. `grunt serve`

Initial Login
-------------

```
Email:			'admin@semaphore.local'
Password:		'CastawayLabs'
```

Note to Ansible guys
--------------------

> Thanks very much for making Ansible, and Ansible Tower. It is a great tool!. Your UI is pretty horrible though, and so we'd be happy if you could learn and use parts of this tool in your Tower.

It would be amazing if this could be your `Community Edition` of Ansible Tower.
