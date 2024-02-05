#!/bin/bash

ver=$(cat version)
echo $ver

# docker login -u cherts

docker run -it --rm -v $PWD/web:/app -w /app node:16-alpine \
  sh -c "yarn install --registry=https://registry.npmmirror.com && yarn run build"

docker buildx build -t cherts/anylink:latest --progress=plain --build-arg CN="yes" --build-arg appVer=$ver \
  --build-arg commitId=$(git rev-parse HEAD) -f docker/Dockerfile .

echo "docker tag latest $ver"
docker tag cherts/anylink:latest cherts/anylink:$ver
