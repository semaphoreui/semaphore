# Semaphore with OpenLDAP example

1. Start stack by command:
   ```
   docker-compose up -d
   ```
2. Create new LDAP user:
   1. Open https://localhost:6443
   2. Login as `cn=admin,dc=example,dc=org` with password `admin`
   3. Create new user `john`
   
      <img src="https://github.com/semaphoreui/semaphore/assets/914224/4eee81d7-0e22-4e20-9bc2-385add519ab5" width="600">

3. Create new Semaphore project:
   1. Open http://localhost:3000
   2. Login as `john`
   3. Create demo project

      <img src="https://github.com/semaphoreui/semaphore/assets/914224/98b780a7-bfbc-4b45-941f-7dd6ca337685" width="600">
