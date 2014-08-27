# HAproxy-rest
---

HAproxy-rest is a REST interface for HAproxy. It exports basic functions normally done through the config file or the
HAproxy's socket interface to a handy REST interface.

## Getting started

Install HAproxy 1.5 or greater in whatever way you like. Just make sure the `haproxy` executable is in your `PATH`. For Ubuntu, use:


    $ add-apt-repository ppa:vbernat/haproxy-1.5 -y  
    $ apt-get update -y  
    $ apt-get install -y haproxy  


Clone this repo 

    git clone https://github.com/tnolet/haproxy-rest 

CD into the directory just created and startup haproxy

OSX:

    $ cd haproxy-rest
    $ haproxy -f resources/haproxy_init.cfg -p resources/haproxy-private.pid -st $(<resources/haproxy-private.pid)

Ubuntu

    $ cd haproxy-rest      
    $ haproxy -f resources/haproxy_init.cfg -p resources/haproxy-private.pid -sf $(cat resources/haproxy-private.pid)

    
Build the program and run it
 
    $ go build
    $ ./haproxy-rest

 

## Parameters

* `-port` Port/IP to use for the REST interface. default: `10001`
* `-configFile` Location of the target HAproxy config file. default: `resources/haproxy_new.cfg`
* `-template` Template file to build HAproxy config. default: `resources/haproxy_cfg.template`
* `-binary` Path to the HAproxy binary. default: `usr/local/bin/haproxy`
* `-pidFile` Location of the HAproxy PID file. default: `resources/haproxy-private.pid`
    
for example, this would startup haproxy-rest on port 12345

    $ ./haproxy-rest -port=12345

## Inspiration

Part of Haproxy-rest is inspired by [haproxy-config](https://github.com/jbuchbinder/haproxy-config) and
[consul-haproxy](https://github.com/hashicorp/consul-haproxy). It is not a straight fork or clone of either of these,
but parts are borrowed.