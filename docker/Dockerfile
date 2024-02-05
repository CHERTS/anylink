#node:16-bullseye
#golang:1.20-bullseye
#debian:bullseye-slim
#sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list

# Use with github action
# You need to compile the ui file first and then execute docker compilation

# server
FROM golang:1.20-alpine3.19 as builder_golang

ARG CN="no"
ARG appVer="appVer"
ARG commitId="commitId"

ENV TZ=Europe/Moscow

WORKDIR /server
COPY docker/init_build.sh /tmp/
COPY server/ /server/
COPY web/ui  /server/ui

#RUN apk add gcc musl-dev bash
RUN sh /tmp/init_build.sh

# anylink
FROM alpine:3.19
LABEL maintainer="github.com/cherts"

ARG CN="no"

ENV TZ=Europe/Moscow
ENV ANYLINK_IN_CONTAINER=true

WORKDIR /app
COPY docker/init_release.sh /tmp/

COPY --from=builder_golang /server/anylink  /app/
COPY docker/docker_entrypoint.sh server/bridge-init.sh ./README.md ./LICENSE version_info /app/
COPY ./deploy /app/deploy
COPY ./index_template  /app/index_template
COPY ./server/conf  /app/conf

#TODO Use mirroring when packaging locally
RUN sh /tmp/init_release.sh

EXPOSE 443 8800 443/udp

#CMD ["/app/anylink"]
ENTRYPOINT ["/app/docker_entrypoint.sh"]
