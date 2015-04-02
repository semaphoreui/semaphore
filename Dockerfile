FROM iojs:onbuild

# node settings
ENV NODE_ENV=production
# starts on port 443 if truthy, else 80
ENV USE_SSL=

# admin user
ENV ADMIN_EMAIL=admin@semaphore.local
ENV ADMIN_USERNAME=semaphore
ENV ADMIN_REALNAME=Administrator
ENV ADMIN_PASSWORD=CastawayLabs

# mongodb
ENV MONGODB_URL mongodb://127.0.0.1/semaphore

# redis config
ENV REDIS_HOST=127.0.0.1
ENV REDIS_PORT=6379
ENV REDIS_KEY=

# smtp config
ENV SMTP_USER=
ENV SMTP_PASS=

# external services
ENV BUGSNAG_KEY=
ENV USE_ANALYTICS=

# add and install
ADD . /srv/semaphore
WORKDIR /srv/semaphore

RUN npm install -g bower
RUN npm install
RUN bower install --allow-root
CMD ["node", "/srv/semaphore/bin/semaphore"]

EXPOSE 80 443
