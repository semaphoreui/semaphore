Docker Compose
##############

To run Semaphore in a simple docker configuration run the following command:

    docker-compose up

You can then access Semaphore directly from the url http://localhost:8081/

SSL Termination Using Nginx
###########################

Generate a cert, ca cert, and key file and place into `Deploy/proxy/cert/` with
these names:

* `cert.pem`
* `privkey.pem`
* `fullchain.pem`

(I've used letsencrypt generated certs with success.)

Run `docker-compose up` and your Semaphore instance will then be at the url
https://localhost:8443

