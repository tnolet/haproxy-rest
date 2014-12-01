#!/bin/bash
set -x

# build the app for linux/i386 an build the docker container
GOOS=linux GOARCH=386 go build
mv ./haproxy-rest target/linux_i386/
rm resources/haproxy_new.cfg
rm resources/haproxy-private.pid
rm resources/persistent_*
docker build -t tnolet/haproxy-rest:$1 .