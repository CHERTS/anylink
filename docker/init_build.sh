#!/bin/sh

set -x

#TODO Use mirroring when packaging locally
if [[ $CN == "yes" ]]; then
  sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
  export GOPROXY=https://goproxy.cn
fi

apk add build-base tzdata gcc g++ musl musl-dev upx

uname -a
env
date

cd /server

go mod tidy

echo "start build"

ldflags="-s -w -X main.appVer=$appVer -X main.commitId=$commitId -X main.buildDate=$(date -Iseconds) -extldflags \"-static\" "

export CGO_ENABLED=1
go build -v -o anylink -trimpath -ldflags "$ldflags"

ls -lh /server/

# Compressed file
upx -9 -k anylink

/server/anylink -v
