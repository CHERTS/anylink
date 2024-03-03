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

cpath=$(pwd)

ver=$(cat version)
echo "Current version: $ver"

echo "Copy binary file"
cd $cpath/server
# -tags osusergo,netgo,sqlite_omit_load_extension
flags="-trimpath"
ldflags="-s -w -extldflags '-static' -X main.appVer=$ver -X main.commitId=$(git rev-parse HEAD) -X main.buildDate=$(date --iso-8601=seconds)"
#github action
gopath=/go

dockercmd=$(
  cat <<EOF
apk add gcc g++ musl musl-dev tzdata
go mod tidy
export CGO_ENABLED=1
go build -v -o anylink_amd64 $flags -ldflags "$ldflags"
./anylink_amd64 -v
EOF
)

# Compile using musl-dev
docker run -q --rm -v $PWD:/app -v $gopath:/go -w /app --platform=linux/amd64 \
  golang:1.20-alpine3.19 sh -c "$dockercmd"

exit 0

# Compile arm64
docker run -q --rm -v $PWD:/app -v $gopath:/go -w /app --platform=linux/arm64 \
  golang:1.20-alpine3.19 go build -o anylink_arm64 $flags -ldflags "$ldflags"
./anylink_arm64 -v

exit 0

cd $cpath

echo "Build deployment files..."
deploy="anylink-deploy"
rm -rf $deploy ${deploy}.tar.gz
mkdir $deploy
mkdir $deploy/log
cp -r server/anylink $deploy
cp -r server/bridge-init.sh $deploy
cp -r server/conf $deploy
cp -r systemd $deploy
cp -r LICENSE $deploy
cp -r home $deploy
tar zcvf ${deploy}.tar.gz $deploy

# Make sure to run with root privileges
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
