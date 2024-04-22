#!/bin/bash

#github action release.sh

set -x
function RETVAL() {
  rt=$1
  if [ $rt != 0 ]; then
    echo $rt
    exit 1
  fi
}

#Current directory
cpath=$(pwd)

ver=$(cat version)
echo "Current version: $ver"

rm -rf artifact-dist
mkdir artifact-dist

function archive() {
  arch=$1
  arch_name=${arch//\//-}
  echo $arch_name
  deploy="anylink-$ver-$arch_name"
  docker container rm $deploy
  docker container create --platform $arch --name $deploy cherts/anylink:$ver
  rm -rf anylink-deploy
  docker cp -a $deploy:/app ./anylink-deploy
  ls -lh anylink-deploy
  tar zcf ${deploy}.tar.gz anylink-deploy
  mv ${deploy}.tar.gz artifact-dist/
}

echo "Copy binary file"

archive "linux/amd64"
archive "linux/arm64"

ls -lh artifact-dist

#Make sure to run with root privileges
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
