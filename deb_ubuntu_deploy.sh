#!/bin/bash

clear

echo " -"
echo " - Ansible Semaphore Install n Deploy - Debian 8.x - Ubuntu 14.04+"
echo " - Semaphore is a Dashboard for Ansible playbook management"
echo " - Warning! Script can only work if executed with root or sudo permissions"
echo " -"

read -p " - Continue (y/n)?" choice

case "$choice" in

  y|Y )
  cd /tmp
  wget https://dev.mysql.com/get/mysql-apt-config_0.8.0-1_all.deb
  dpkg -i mysql-apt-config_0.8.0-1_all.deb
  apt-get update
  rm /tmp/mysql-apt-config_0.8.0-1_all.deb
  cd ~/
  apt-get install -y build-essential curl ansible git mysql-community-server
  apt-get install -f -y
  curl -L https://github.com/ansible-semaphore/semaphore/releases/download/v2.0.4/semaphore_linux_amd64 > /usr/bin/semaphore
  chmod u+x /usr/bin/semaphore
  semaphore -setup
  echo " -"
  read -p " - Give Semaphore Playbook Folder again (for default enter: /tmp/semaphore):" playbook
  echo  " -"
  echo  " - Smooth and fine"
  echo " -"
  read -p " - Ready to Deploy on http://localhost:3000 (y/n)?" choice2

     case "$choice2" in
        y|Y )
        semaphore -config $playbook/semaphore_config.json;;
        n|N )
        echo  " -"
        echo  " - Bye Bye"
        echo  " -";;
     esac;;

  n|N )
  echo  " -"
  echo  " - Bye Bye"
  echo  " -";;

  * )
  echo  " -"
  echo  " - Invalid choice, please re-run the script"
  echo  " -";;

esac
