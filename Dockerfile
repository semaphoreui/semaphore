FROM iojs:latest

RUN ln -snf /usr/bin/nodejs /usr/bin/node
RUN apt-get update && apt-get install -y python-dev build-essential python-pip git && pip install ansible && apt-get clean

ADD . /srv/semaphore
WORKDIR /srv/semaphore

RUN npm install
RUN ./node_modules/.bin/bower install --allow-root

ENV NODE_ENV production
ENV REDIS_PORT 6379
CMD ["node", "/srv/semaphore/bin/semaphore"]

EXPOSE 80
