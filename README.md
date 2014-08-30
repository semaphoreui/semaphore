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

Install requirements:
- node.js >= 0.11.x
- an isolated environment (e.g. Docker / NodeGear)
- ansible (the tool)
- mongodb & redis
- Sudo access (this might change). To run jobs, this tool writes private keys to /root/.ssh and copies playbook directories to /root/.

1. `cp lib/credentials.example.json lib/credentials.json` (<- make custom to your environment)
2. `bower install`
3. `grunt serve`

Open [localhost:3000](http://localhost:3000)

Initial Login
-------------

```
Email:			'admin@semaphore.local'
Password:		'CastawayLabs'
```

Runs Best on [NodeGear](https://nodegear.com)
---------------------

Semaphore is used internally at NodeGear, and at CastawayLabs. We've made this because we think Ansible Tower is _way_ too expensive.

Note to Ansible guys
--------------------

> Thanks very much for making Ansible, and Ansible Tower. It is a great tool!. Your UI is pretty horrible though, and so we'd be happy if you could learn and use parts of this tool in your Tower.

It would be amazing if this could be your `Community Edition` of Ansible Tower.

Acknowledgments
---------------

This product was hacked together in ~2 days (just about ~~16~~ 20 hours). We're sorry if parts of it don't work, or are unusable. It isn't 100% flexible and bug-free.

If you know Angular.js, UI-Router and node.js, we'd be happy for you to help out. Make an issue if you would like something to work on :P.

**Todo List**

1. User Management
2. Add login rate limits
3. More options for running jobs/playbooks
4. Nicer UI