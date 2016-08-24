FROM fedora

RUN dnf -y install git ansible mariadb; dnf -y clean all

ENV SEMAPHORE_DB_USER semaphore
ENV SEMAPHORE_DB_PASS semaphore
ENV SEMAPHORE_DB_HOST mysql
ENV SEMAPHORE_DB_PORT 3306
ENV SEMAPHORE_DB semaphore
ENV SEMAPHORE_PLAYBOOK_PATH /etc/semaphore/
ENV SEMAPHORE_ADMIN_PASSWORD cangetin
ENV SEMAPHORE_ADMIN_NAME Default\ Administrator
ENV SEMAPHORE_ADMIN_EMAIL admin@localhost
ENV SEMAPHORE_ADMIN admin

RUN curl -L https://github.com/ansible-semaphore/semaphore/releases/download/v2.0.4/semaphore_linux_amd64 > /usr/bin/semaphore && chmod +x /usr/bin/semaphore && mkdir -p /etc/semaphore/playbooks

ADD semaphore-startup.sh /usr/bin/semaphore-startup.sh

RUN chmod +x /usr/bin/semaphore-startup.sh

EXPOSE 3000

ENTRYPOINT ["/usr/bin/semaphore-startup.sh"]

CMD ["/usr/bin/semaphore", "-config", "/etc/semaphore/semaphore_config.json"]
