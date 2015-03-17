FROM iojs:onbuild

ENV NODE_ENV production

ADD . /srv/semaphore
WORKDIR /srv/semaphore

RUN npm install
CMD ["node", "/srv/semaphore/bin/semaphore"]

EXPOSE 80