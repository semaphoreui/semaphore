# Ansible Semaphore

[![semaphore](https://snapcraft.io/semaphore/badge.svg)](https://snapcraft.io/semaphore)
[![Join the chat at https://gitter.im/AnsibleSemaphore/semaphore](https://img.shields.io/gitter/room/AnsibleSemaphore/semaphore?logo=gitter)](https://gitter.im/AnsibleSemaphore/semaphore?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![Twitter](https://img.shields.io/twitter/follow/semaphoreui?style=social&logo=twitter)](https://twitter.com/semaphoreui)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/fiftin)

Ansible Semaphore is a modern UI for Ansible. It lets you easily run Ansible playbooks, get notifications about fails, control access to deployment system.

If your project has grown and deploying from the terminal is no longer for you then Ansible Semaphore is what you need.

![responsive-ui-phone1](https://user-images.githubusercontent.com/914224/134777345-8789d9e4-ff0d-439c-b80e-ddc56b74fcee.png)

## Installation

### Full documentation
https://docs.ansible-semaphore.com/administration-guide/installation

### Snap

```bash
sudo snap install semaphore
sudo semaphore user add --admin --name "Your Name" --login your_login --email your-email@examaple.com --password your_password
```
[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/semaphore)

### Docker 

https://hub.docker.com/r/semaphoreui/semaphore

`docker-compose.yml` for minimal configuration:

```yaml
services:
  semaphore:
    ports:
      - 3000:3000
    image: semaphoreui/semaphore:latest
    environment:
      SEMAPHORE_DB_DIALECT: bolt
      SEMAPHORE_ADMIN_PASSWORD: changeme
      SEMAPHORE_ADMIN_NAME: admin
      SEMAPHORE_ADMIN_EMAIL: admin@localhost
      SEMAPHORE_ADMIN: admin
    volumes:
      - /path/to/data/home:/etc/semaphore # config.json location
      - /path/to/data/lib:/var/lib/semaphore # database.boltdb location (Not required if using mysql or postgres)
```
### Kubernetes
Quick start with kubernetes:


```yaml
apiVersion: v1
kind: Secret
metadata:
  name: semaphore
type: Opaque
data:
  # echo -n 'admin' | base64
  semaphore.admin.name: YWRtaW4=
  # echo -n 'changeme' | base64
  semaphore.admin.password: Y2hhbmdlbWU=
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: semaphore
  labels:
    app: semaphore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: semaphore
  template:
    metadata:
      name: semaphore
      labels:
        app: semaphore
    spec:
      containers:
        - name: semaphore
          image: semaphoreui/semaphore:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3000
          env:
            - name: SEMAPHORE_DB_DIALECT
              value: bolt
            - name: SEMAPHORE_ADMIN_EMAIL
              value: admin@localhost
            - name: SEMAPHORE_ADMIN
              value: admin
            - name: SEMAPHORE_ADMIN_NAME
              valueFrom:
                secretKeyRef:
                  name: semaphore
                  key: semaphore.admin.name
            - name: SEMAPHORE_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: semaphore
                  key: semaphore.admin.password
          volumeMounts:
            - mountPath: /var/lib/semaphore
              name: bolt
      volumes:
        # volume to store semaphore data, otherwise it will be lost after pod restart
        - name: bolt
          emptyDir: {}
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: semaphore
spec:
  selector:
    app: semaphore
  ports:
    - protocol: TCP
      port: 80
      targetPort: 3000
  # you can choose NodePort instead of ClusterIP to expose the service
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: semaphore
spec:
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: semaphore
                port:
                  number: 80
      # use your own domain name
      host: semaphore.localhost.com
```

Use `config.json` to configure Semaphore
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: semaphore
type: Opaque
data:
  # echo -n 'admin' | base64
  semaphore.admin.name: YWRtaW4=
  # echo -n 'changeme' | base64
  semaphore.admin.password: Y2hhbmdlbWU=
  # base config file, you can add your own config here
  # base64 -i config.json | fold -w 76
  config.json: |
    ewogICJib2x0IjogewogICAgImhvc3QiOiAiL2hvbWUvdWJ1bnR1L3NlbWFwaG9yZS5ib2x0Igog
    IH0sCiAgIm15c3FsIjogewogICAgImhvc3QiOiAibG9jYWxob3N0IiwKICAgICJ1c2VyIjogInJv
    b3QiLAogICAgInBhc3MiOiAiKioqKioiLAogICAgIm5hbWUiOiAic2VtYXBob3JlIiwKICAgICJv
    cHRpb25zIjoge30KICB9LAogICJwb3N0Z3JlcyI6IHsKICAgICJob3N0IjogImxvY2FsaG9zdCIs
    CiAgICAidXNlciI6ICJwb3N0Z3JlcyIsCiAgICAicGFzcyI6ICIqKioqKiIsCiAgICAibmFtZSI6
    ICJzZW1hcGhvcmUiLAogICAgIm9wdGlvbnMiOiB7fQogIH0sCiAgImRpYWxlY3QiOiAiYm9sdCIs
    CiAgInBvcnQiOiAiIiwKICAiaW50ZXJmYWNlIjogIiIsCiAgInRtcF9wYXRoIjogIi90bXAvc2Vt
    YXBob3JlIiwKICAiY29va2llX2hhc2giOiAiKioqKioiLAogICJjb29raWVfZW5jcnlwdGlvbiI6
    ICIqKioqKiIsCiAgImFjY2Vzc19rZXlfZW5jcnlwdGlvbiI6ICIqKioqKiIsCiAgImVtYWlsX3Nl
    bmRlciI6ICIiLAogICJlbWFpbF9ob3N0IjogIiIsCiAgImVtYWlsX3BvcnQiOiAiIiwKICAid2Vi
    X2hvc3QiOiAiIiwKICAibGRhcF9iaW5kZG4iOiAiIiwKICAibGRhcF9iaW5kcGFzc3dvcmQiOiAi
    IiwKICAibGRhcF9zZXJ2ZXIiOiAiIiwKICAibGRhcF9zZWFyY2hkbiI6ICIiLAogICJsZGFwX3Nl
    YXJjaGZpbHRlciI6ICIiLAogICJsZGFwX21hcHBpbmdzIjogewogICAgImRuIjogIiIsCiAgICAi
    bWFpbCI6ICIiLAogICAgInVpZCI6ICIiLAogICAgImNuIjogIiIKICB9LAogICJ0ZWxlZ3JhbV9j
    aGF0IjogIiIsCiAgInRlbGVncmFtX3Rva2VuIjogIiIsCiAgImNvbmN1cnJlbmN5X21vZGUiOiAi
    IiwKICAibWF4X3BhcmFsbGVsX3Rhc2tzIjogMCwKICAiZW1haWxfYWxlcnQiOiBmYWxzZSwKICAi
    dGVsZWdyYW1fYWxlcnQiOiBmYWxzZSwKICAic2xhY2tfYWxlcnQiOiBmYWxzZSwKICAibGRhcF9l
    bmFibGUiOiBmYWxzZSwKICAibGRhcF9uZWVkdGxzIjogZmFsc2UKfQ==
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: semaphore
  labels:
    app: semaphore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: semaphore
  template:
    metadata:
      name: semaphore
      labels:
        app: semaphore
    spec:
      containers:
        - name: semaphore
          image: semaphoreui/semaphore:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 3000
          env:
            - name: SEMAPHORE_DB_DIALECT
              value: bolt
            - name: SEMAPHORE_ADMIN_EMAIL
              value: admin@localhost
            - name: SEMAPHORE_ADMIN
              value: admin
            - name: SEMAPHORE_ADMIN_NAME
              valueFrom:
                secretKeyRef:
                  name: semaphore
                  key: semaphore.admin.name
            - name: SEMAPHORE_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: semaphore
                  key: semaphore.admin.password
          volumeMounts:
            - mountPath: /var/lib/semaphore
              name: bolt
            - name: config
              mountPath: /etc/semaphore
              readOnly: true
      volumes:
        # volume to store semaphore data, otherwise it will be lost after pod restart
        - name: bolt
          emptyDir: {}
        # Mount 'config.json' file from the Secret into the Pod
        - name: config
          secret:
            secretName: semaphore
            items:
              - key: config.json
                path: config.json
      restartPolicy: Always
---
apiVersion: v1
kind: Service
metadata:
  name: semaphore
spec:
  selector:
    app: semaphore
  ports:
    - protocol: TCP
      port: 80
      targetPort: 3000
  # you can choose NodePort instead of ClusterIP to expose the service
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: semaphore
spec:
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: semaphore
                port:
                  number: 80
      # use your own domain name
      host: semaphore.localhost.com
```

## Demo

You can test latest version of Semaphore on https://demo.ansible-semaphore.com.

## Docs

Admin and user docs: https://docs.ansible-semaphore.com

API description: https://ansible-semaphore.com/api-docs/

## Contributing

If you want to write an article about Ansible or Semaphore, contact [@fiftin](https://github.com/fiftin) and we will place your article in our [Blog](https://www.ansible-semaphore.com/blog/) with link to your profile.

PR's & UX reviews are welcome!

Please follow the [contribution](https://github.com/ansible-semaphore/semaphore/blob/develop/CONTRIBUTING.md) guide. Any questions, please open an issue.

## Release Signing

All releases after 2.5.1 are signed with the gpg public key
`8CDE D132 5E96 F1D9 EABF 17D4 2C96 CF7D D27F AB82`

## Support

If you like Ansible Semaphore, you can support the project development on [Ko-fi](https://ko-fi.com/fiftin).

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
