# Semaphore with Authentik LDAP example


1. Start stack by command:
   ```
   echo "PG_PASS=$(openssl rand 36 | base64)" >> .env
   echo "AUTHENTIK_SECRET_KEY=$(openssl rand 60 | base64)" >> .env
   docker-compose up -d
   ```
2. To start the initial setup, navigate to http://localhsot:9000/if/flow/initial-setup/. 
   There you are prompted to set a password for the `akadmin` user (the default user).
3. Set up LDAP in Athentik according to the video tutorial [Authentik - LDAP Generic Setup](https://youtu.be/RtPKMMKRT_E).
4. Set up Athentik LDAP and Semaphore containers:
   1. Copy `AUTHENTIK_TOKEN`.
   2. Update `AUTHENTIK_TOKEN` for LDAP service.
   3. Reload the stack:
      ```
      docker-compose down
      docker-compose up -d
      ```
5. Create new Semaphore project:
    1. Open http://localhost:3000
    2. Login as `john`
    3. Create demo project

       <img src="https://github.com/semaphoreui/semaphore/assets/914224/98b780a7-bfbc-4b45-941f-7dd6ca337685" width="600">
