#!/bin/bash

printf "#\041/bin/bash\nssh -i /root/.ssh/id_rsa \$1 \$2\n" > /root/ssh_wrapper.sh
chmod +x /root/ssh_wrapper.sh

cd /root
GIT_SSH=/root/ssh_wrapper.sh git clone "$1" $2
if [ $? -ne 0 ]; then
	exit 2
fi