task compile

GOOS=linux GOARCH=amd64 task build:local

scp ./bin/semaphore hetzner:~/

#ssh hetzner << EOF
#sudo systemctl stop demo2.semaphore
#sudo cp ~/semaphore /usr/bin/semaphore
#sudo systemctl start demo2.semaphore
#EOF