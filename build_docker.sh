#!/bin/bash

ver=$(cat version)
echo "Current version: $ver"

#docker login -u cherts

echo "Docker build..."
docker build -t cherts/anylink:latest --progress=plain --build-arg appVer=$ver \
  --build-arg commitId=$(git rev-parse HEAD) -f docker/Dockerfile .

echo "Docker tag latest $ver"
docker tag cherts/anylink:latest cherts/anylink:$ver

#echo "Docker push..."
#docker push cherts/anylink
