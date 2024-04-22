#!/bin/bash

echo "Build deployment files..."
deploy="anylink-deploy"
rm -rf $deploy ${deploy}.tar.gz
mkdir $deploy
mkdir $deploy/log
mkdir $deploy/systemd
cp -r server/anylink $deploy
cp -r server/bridge-init.sh $deploy
cp -r server/conf $deploy
cp -r deploy/anylink.service $deploy/systemd
cp -r LICENSE $deploy
cp -r index_template $deploy
tar zcvf ${deploy}.tar.gz $deploy

# Make sure to run with root privileges
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
