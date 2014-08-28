FROM ubuntu:latest

MAINTAINER tim@magnetic.io

# This Dockerfile does the basic install of Haproxy and Haproxy-rest. Please see:
# https://github.com/tnolet/haproxy-rest
#
# HAproxy is currently version 1.5.3 build from source on Ubuntu with the following options
# apt-get install build-essential
# apt-get install libpcre3-dev
# make TARGET=linux26 ARCH=i386 USE_PCRE=1 USE_LINUX_SPLICE=1 USE_LINUX_TPROXY=1
#
#

ADD ./target/linux_i386/haproxy-rest /haproxy-rest

ADD ./resources /resources

ADD ./dependencies/haproxy /usr/local/bin/haproxy

EXPOSE 80

EXPOSE 10001

EXPOSE 1988

ENTRYPOINT ["/haproxy-rest"]

