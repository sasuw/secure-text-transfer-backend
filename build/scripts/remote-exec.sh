 #!/bin/bash

#TO BE EXECUTED ON THE REMOTE CORE OS INSTANCE

DOCKER_CONTAINER_NAME="stt-backend"
DOCKER_IMAGE_NAME="stt-backend-small"
DOCKER_FILE_NAME=$DOCKER_IMAGE_NAME

sudo -s
bzip2 -d $DOCKER_FILE_NAME_.tar.bz2
docker load -i $DOCKER_FILE_NAME.tar
docker ps -q --filter "name=$DOCKER_CONTAINER_NAME" | grep -q . && docker stop $DOCKER_CONTAINER_NAME && docker rm -fv $DOCKER_CONTAINER_NAME
docker run -d -p 9999:9999 --name stt-backend sasu/$DOCKER_IMAGE_NAME
rm $DOCKER_FILE_NAME.tar*