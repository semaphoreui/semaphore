#!/bin/bash

printf "#\041/bin/bash\nssh -i \"$HOME/.ssh/id_rsa\" \$1 \$2\n" > $HOME/ssh_wrapper.sh
chmod +x $HOME/ssh_wrapper.sh

cd $HOME
GIT_SSH=$HOME/ssh_wrapper.sh git clone "$1" $2
if [ $? -ne 0 ]; then
	exit 2
fi