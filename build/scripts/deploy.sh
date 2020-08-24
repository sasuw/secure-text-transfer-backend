#!/bin/bash

SSH_PORT="22"
SSH_HOST="backend.stt.sasu.net"

#SSH_PORT="2222"
#SSH_HOST="127.0.0.1"

mkdir -p tmp
cd tmp
docker save -o stt-backend-small.tar sasu/stt-backend-small
bzip2 -z stt-backend-small.tar
scp -P $SSH_PORT stt-backend-small.tar.bz2 core@$SSH_HOST:/var/home/core
cat ../build/scripts/remote-exec.sh | ssh $SSH_HOST -p $SSH_PORT -l core
rm stt-backend-small.tar.bz2