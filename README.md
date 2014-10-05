# HAproxy-rest
---

HAproxy-rest is a REST interface for HAproxy. It exports basic functions normally done through the config file or the
HAproxy's socket interface to a handy REST interface. Features are:

-   Update the config through REST
-   Adjust server weight
-   Get statistics on frontends, backends and servers
-   Stream statistics to Kafka

*Important* : Currently, HAproxy-rest does NOT check validity of the HAproxy commands and configs submitted to it.
Submitting a config where a frontend references a non-existing backend will be accepted by the REST api but crash HAproxy

## Installing: the easy Docker way

Start up an instance with all defaults and bind it to the local network interface

    $ docker run --net=host tnolet/haproxy-rest


The default ports are 1988 for stats, 10001 for the REST api and 80 for the frontend. You should see syslog getting started
and the [gin-gonic](https://github.com/gin-gonic/gin) web framework spit out all the know routes.
 
    $ docker run --net=host tnolet/haproxy-rest
    2014-10-05 12:44:14 INFO  HaproxyReload: 16552
    2014-10-05 12:44:14 INFO  Connecting to Kafka on localhost:9092...
    2014-10-05 12:44:14 INFO  Connection to Kafka successful
    2014-10-05 12:44:14 INFO  Starting REST server
    [GIN-debug] POST  /v1/backend/:name/:server/weight/:weight --> main.func·002 (4 handlers)
    [GIN-debug] GET   /v1/stats                 --> main.func·003 (4 handlers)
    [GIN-debug] GET   /v1/stats/backend         --> main.func·004 (4 handlers)
    [GIN-debug] GET   /v1/stats/frontend        --> main.func·005 (4 handlers)
    [GIN-debug] GET   /v1/stats/server          --> main.func·006 (4 handlers)
    [GIN-debug] GET   /v1/config                --> main.func·007 (4 handlers)
    [GIN-debug] POST  /v1/config                --> main.func·008 (4 handlers)
    [GIN-debug] GET   /v1/info                  --> main.func·009 (4 handlers)
    [GIN-debug] Listening and serving HTTP on 0.0.0.0:10001
### Changing the port

You could change the REST api port by adding the `-port` flag

    $ docker run --net=host tnolet/haproxy-rest -port=1234

Or by exporting an environment variable `PORT0`. When deploying with Marathon 0.7.0, this is done automatically
     
     $ export PORT0=12345
     $ docker run --net=host tnolet/haproxy-rest

### Getting statistics/metrics

Statistics are published in two different ways: straight from the REST interface and as Kafka topics

## Stats via REST
     
Grab some stats from the `/stats` endpoint. Notice the IP address. This is [boot2docker](https://github.com/boot2docker/boot2docker)'s address on my Macbook. I'm using [httpie](https://github.com/jakubroztocil/httpie) instead of Curl.

    $ http http://192.168.59.103:10001/v1/stats
    HTTP/1.1 200 OK
    
    [
        {
            "act": "", 
            "bck": "", 
            "bin": "3572", 
            "bout": "145426", 
            "check_code": "", 
            "check_duration": "", 
            "check_status": "", 
            "chkdown": "", 
            "chkfail": "", 
            "cli_abrt": "", 
            ...
            
Valid endpoints are `stats/frontend`, `stats/backend` and `stats/server`. The `/stats` endpoint gives you all of them
in one go.

## Stats via Kafka

Statistics are also published as Kafka topics. Configure a Kafka endpoint using the `-kakfaHost` and `-kafkaPort` flags.
Stats are published as the following topics

- loadbalancer.frontend
- loadbalancer.backend
- loadbalancer.server

            
### Updating the configuration

Post a configuration. You can use the example file `resources/config_example.json`

    $ http POST http://192.168.59.103:10001/v1/config < resources/config_example.json 
    HTTP/1.1 200 OK
     
    Ok
    
Update the weight of some backend server

    $ http POST http://192.168.59.103:10001/v1/backend/testbe/test_be_1/weight/20
    HTTP/1.1 200 OK
    
    Ok

    
## Installing: the harder custom build way

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

    
Build the program and run it. 
 
    $ go build
    $ ./haproxy-rest

If you're on Mac OSX or Windows and want to compile for Linux (which is probably the OS 
you're using to run HAproxy), you need to cross compile. 
For this, go to your Go `src` directory, i.e.

    $ cd /usr/local/Cellar/go/1.3.1

Compile the compiler with the correct arguments for OS and ARC

    $ GOOS=linux GOARCH=386 CGO_ENABLED=0 ./make.bash --no-clean

Compile the application

    $ GOOS=windows GOARCH=386 go build 
 

## Parameters

* `-port` Port/IP to use for the REST interface. default: `10001`
* `-configFile` Location of the target HAproxy config file. default: `resources/haproxy_new.cfg`
* `-template` Template file to build HAproxy config. default: `resources/haproxy_cfg.template`
* `-binary` Path to the HAproxy binary. default: `/usr/local/bin/haproxy`
* `-kafkaHost` The hostname or ip address of the Kafka host. default: `localhost`
* `-kafkaPort` The port of the Kafka host. default: `9092`
* `-pidFile` Location of the HAproxy PID file. default: `resources/haproxy-private.pid`
    
for example, this would startup haproxy-rest on port 12345

    $ ./haproxy-rest -port=12345

## Inspiration

Part of Haproxy-rest is inspired by [haproxy-config](https://github.com/jbuchbinder/haproxy-config) and
[consul-haproxy](https://github.com/hashicorp/consul-haproxy). It is not a straight fork or clone of either of these,
but parts are borrowed.