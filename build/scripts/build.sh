#!/bin/bash

cd cmd
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
cd ..
docker build -t sasu/stt-backend-small -f build/Dockerfile.scratch .