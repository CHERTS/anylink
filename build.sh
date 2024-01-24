#!/bin/bash

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

ver=`cat server/base/app_ver.go | grep APP_VER | awk '{print $3}' | sed 's/"//g'`
echo "Version: $ver"

echo "Compile frontend project..."
cd $cpath/web
#Domestic alternative sources speed up
#npx browserslist@latest --update-db
#npm install --registry=https://registry.npm.taobao.org
#npm install
#npm run build

yarn install --registry=https://registry.npmmirror.com
yarn run build


RETVAL $?

echo "Compile binaries..."
cd $cpath/server
rm -rf ui
cp -rf $cpath/web/ui .
#Domestic alternative sources speed up
#export GOPROXY=https://goproxy.io
go mod tidy
go build -v -o anylink -ldflags "-s -w -X main.CommitId=$(git rev-parse HEAD)"
RETVAL $?

cd $cpath

echo "Organize deployment files..."
deploy="anylink-deploy"
rm -rf $deploy ${deploy}.tar.gz
mkdir $deploy

cp -r server/anylink $deploy
#cp -r server/bridge-init.sh $deploy
cp -r server/conf $deploy

cp -r systemd $deploy
cp -r LICENSE $deploy
cp -r home $deploy

tar zcvf ${deploy}.tar.gz $deploy

#Make sure to run with root privileges
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
