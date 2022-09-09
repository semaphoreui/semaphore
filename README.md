# Ansible Semaphore

[![Circle CI](https://circleci.com/gh/ansible-semaphore/semaphore.svg?style=svg&circle-token=3702872acf2bec629017fa7dd99fdbea56aef7df)](https://circleci.com/gh/ansible-semaphore/semaphore)
[![Twitter](https://img.shields.io/twitter/follow/AnsibleSem?style=social&logo=twitter)](https://twitter.com/AnsibleSem)
[![Snap](https://img.shields.io/badge/snap-semaphore-005c63)](https://snapcraft.io/semaphore)
[![StackShare](https://img.shields.io/badge/tech-stack-008ff9)](https://stackshare.io/ansible-semaphore)
[![Join the chat at https://gitter.im/AnsibleSemaphore/semaphore](https://img.shields.io/gitter/room/AnsibleSemaphore/semaphore?logo=gitter)](https://gitter.im/AnsibleSemaphore/semaphore?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

<!-- [![Release](https://img.shields.io/github/v/release/ansible-semaphore/semaphore.svg)](https://stackshare.io/ansible-semaphore) -->
<!-- [![Godoc Reference](https://pkg.go.dev/badge/github.com/ansible-semaphore/semaphore?utm_source=godoc)](https://godoc.org/github.com/ansible-semaphore/semaphore) -->
<!-- [![Codacy Badge](https://api.codacy.com/project/badge/Grade/89e0129c6ba64fe2b1ebe983f72a4eff)](https://www.codacy.com/app/ansible-semaphore/semaphore?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=ansible-semaphore/semaphore&amp;utm_campaign=Badge_Grade)
[![Codacy Badge](https://api.codacy.com/project/badge/Coverage/89e0129c6ba64fe2b1ebe983f72a4eff)](https://www.codacy.com/app/ansible-semaphore/semaphore?utm_source=github.com&utm_medium=referral&utm_content=ansible-semaphore/semaphore&utm_campaign=Badge_Coverage) -->

Ansible Semaphore is a modern UI for Ansible. It lets you easily run Ansible playbooks, get notifications about fails, control access to deployment system.

If your project has grown and deploying from the terminal is no longer for you then Ansible Semaphore is what you need.

Follow Semaphore on Twitter ([AnsibleSem](https://twitter.com/AnsibleSem)) and StackShare ([ansible-semaphore](https://stackshare.io/ansible-semaphore)).

![responsive-ui-phone1](https://user-images.githubusercontent.com/914224/134777345-8789d9e4-ff0d-439c-b80e-ddc56b74fcee.png)


[![Star History Chart](https://api.star-history.com/svg?repos=ansible-semaphore/semaphore,rundeck/rundeck&type=Date)](https://star-history.com/#ansible-semaphore/semaphore,rundeck/rundeck&Date)


<!--
![image](https://user-images.githubusercontent.com/914224/134411082-48235676-06d2-4d4b-b674-4ffe1e8d0d0d.png)

![semaphore](https://user-images.githubusercontent.com/914224/125253358-c214ed80-e312-11eb-952e-d96a1eba93f6.png)
-->


<!--
- [Releases](https://github.com/ansible-semaphore/semaphore/releases)
- [Installation](https://docs.ansible-semaphore.com/administration-guide/installation)
- [Docker Hub](https://hub.docker.com/r/semaphoreui/semaphore/)
- [Contribution](https://github.com/ansible-semaphore/semaphore/blob/develop/CONTRIBUTING.md)
- [Troubleshooting](https://github.com/ansible-semaphore/semaphore/wiki/Troubleshooting)
- [Roadmap](https://github.com/ansible-semaphore/semaphore/projects)
- [UI Walkthrough](https://blog.strangeman.info/ansible/2017/08/05/semaphore-ui-guide.html) (external blog)
-->

## Installation

### Full documentation
https://docs.ansible-semaphore.com/administration-guide/installation

### Snap

```
sudo snap stop semaphore
sudo semaphore user add --admin --name "Your Name" --login your_login --email your-email@examaple.com --password your_password
sudo snap start semaphore
```
[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/semaphore)

### Docker 

```
docker pull semaphoreui/semaphore
```
https://hub.docker.com/r/semaphoreui/semaphore

## Demo

You can test latest version of Semaphore on https://demo.ansible-semaphore.com.

## Docs

Admin and user docs: https://docs.ansible-semaphore.com

API description: https://ansible-semaphore.com/api/

## Contributing

PR's & UX reviews are welcome!

Please follow the [contribution](https://github.com/ansible-semaphore/semaphore/blob/develop/CONTRIBUTING.md) guide. Any questions, please open an issue.

## Release Signing

All releases after 2.5.1 are signed with the gpg public key
`8CDE D132 5E96 F1D9 EABF 17D4 2C96 CF7D D27F AB82`

## License

MIT License

Copyright (c) 2016 Castaway Consulting LLC

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
