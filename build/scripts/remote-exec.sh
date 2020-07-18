 #!/bin/bash

#TO BE EXECUTED ON THE REMOTE CORE OS INSTANCE
sudo -s
bzip2 -d stt-backend-small.tar.bz2
docker load -i stt-backend-small.tar
docker stop stt-backend
docker rm stt-backend
docker run -d -p 9999:9999 --name stt-backend sasu/stt-backend-small
rm stt-backend-small.tar*