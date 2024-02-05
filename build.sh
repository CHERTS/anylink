#!/bin/bash

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
echo "Current version $ver"

echo "Compile frontend project..."
cd $cpath/web

#npx browserslist@latest --update-db
yarn install --registry=https://registry.npmmirror.com
yarn run build
RETVAL $?

echo "Compile binaries..."
cd $cpath/server
rm -rf ui
cp -rf $cpath/web/ui .

# -tags osusergo,netgo,sqlite_omit_load_extension
flags="-v -trimpath"

# -extldflags '-static'
ldflags="-s -w -X main.appVer=$ver -X main.commitId=$(git rev-parse HEAD) -X main.date=$(date -Iseconds)"

export GOPROXY=https://goproxy.io
go mod tidy
go build -o anylink $flags -ldflags "$ldflags"

cd $cpath

exit 0

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
