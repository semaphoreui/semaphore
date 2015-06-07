#!/bin/bash

printf "#\041/bin/bash\nssh -i /root/.ssh/$3 \$1 \$2\n" > /root/ssh_wrapper_$3.sh
chmod +x /root/ssh_wrapper_$3.sh

cd /root
GIT_SSH=/root/ssh_wrapper_$3.sh git clone "$1" $2
if [ $? -ne 0 ]; then
	exit 2
fi