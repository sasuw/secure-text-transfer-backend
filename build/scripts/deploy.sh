#!/bin/bash

docker save -o stt-backend-small.tar sasu/stt-backend-small
bzip2 -z stt-backend-small.tar
scp -P 2222 stt-backend-small.tar.bz2 core@127.0.0.1:/var/home/core
cat remote-exec.sh | ssh 127.0.0.1 -p 2222 -l core
rm stt-backend-small.tar.bz2