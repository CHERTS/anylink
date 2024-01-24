#!/bin/bash

ver=`cat server/base/app_ver.go | grep APP_VER | awk '{print $3}' | sed 's/"//g'`
echo $ver

#docker login -u cherts

#docker build -t cherts/anylink .

docker build -t cherts/anylink --build-arg GitCommitId=$(git rev-parse HEAD) -f docker/Dockerfile .

docker tag cherts/anylink:latest cherts/anylink:$ver

exit 0

docker push cherts/anylink:$ver
docker push cherts/anylink:latest

