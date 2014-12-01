# HAproxy-rest
---

HAproxy-rest started as a REST interface for HAproxy. Now it's much more. Features are:

-   Update the config through REST or through Zookeeper
-   Run in full load balancer mode, or simple local proxy mode
-   Adjust server weight
-   Get statistics on frontends, backends and servers
-   Stream statistics to Kafka
-   Set ACL's *(experimental)*
-   Set HTTP & TCP Spike limiting *(experimental)*


*Important* : Currently, HAproxy-rest does NOT check validity of the HAproxy command, ACLs and configs submitted to it.
Submitting a config where a frontend references a non-existing backend will be accepted by the REST api but crash HAproxy

### Installing: the easy Docker way

Start up an instance with all defaults and bind it to the local network interface

    $ docker run --net=host tnolet/haproxy-rest

    2014-11-20 21:04:31 INFO   ==> Starting in Load Balancer mode <==
    2014-11-20 21:04:31 WARN  Unable to load persistent config from disk
    2014-11-20 21:04:31 WARN  Loading example config
    2014-11-20 21:04:31 INFO  Pid file exists, proceeding with startup...
    2014-11-20 21:04:31 INFO  HaproxyReload: 30300
    2014-11-20 21:04:31 INFO  Starting REST server
    [GIN-debug] POST  /v1/backend/:name/servers/:server/weight/:weight --> main.func·003 (4 handlers)
    [GIN-debug] POST  /v1/frontend/:name/acls/:acl/:pattern --> main.func·004 (4 handlers)
    [GIN-debug] GET   /v1/frontend/:name/acls   --> main.func·005 (4 handlers)
    [GIN-debug] GET   /v1/stats                 --> main.func·006 (4 handlers)
    [GIN-debug] GET   /v1/stats/backend         --> main.func·007 (4 handlers)
    [GIN-debug] GET   /v1/stats/frontend        --> main.func·008 (4 handlers)
    [GIN-debug] GET   /v1/stats/server          --> main.func·009 (4 handlers)
    [GIN-debug] GET   /v1/config                --> main.func·010 (4 handlers)
    [GIN-debug] POST  /v1/config                --> main.func·011 (4 handlers)
    [GIN-debug] GET   /v1/info                  --> main.func·012 (4 handlers)
    [GIN-debug] Listening and serving HTTP on 0.0.0.0:10001
    
The default ports are:

    10001      REST Api (for config, stats etc)  
    1988       built-in Haproxy stats
    
### Changing ports

You could change the REST api port by adding the `-port` flag

    $ docker run --net=host tnolet/haproxy-rest -port=1234

Or by exporting an environment variable `PORT0`. When deploying with Marathon 0.7.0, this is done automatically
     
     $ export PORT0=12345
     $ docker run --net=host tnolet/haproxy-rest

## Getting statistics

Statistics are published in two different ways: straight from the REST interface and as Kafka topics

### Stats via REST
     
Grab some stats from the `/stats` endpoint. Notice the IP address. This is [boot2docker](https://github.com/boot2docker/boot2docker)'s address on my Macbook. I'm using [httpie](https://github.com/jakubroztocil/httpie) instead of curl.

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

### Stats via Kafka

Statistics are also published as Kafka topics. Configure a Kafka endpoint using the `-kakfaHost` and `-kafkaPort` flags.
Stats are published as the following topic:

- loadbalancer.all

The messages on that topic are json strings, where the "name" key indicates what metric type from which proxy
 you are dealing with, i.e.:

    {
     "name": "testbe.test_be_1.rate",   # The rate for server test_be_1 for proxy testbe
     "value": "2",                      # The value of the metric
     "timestamp": 1413546338            # The timestamp in Unix epoch
    }
    {
     "name": "testbe.test_be_1.rate_lim",
     "value": "12",
     "timestamp": 1413546338
    }
    { "name": "testbe.test_be_1.rate_max",
     "value": "30",
     "timestamp": 1413546338
    }

__Note:__ currently, not all Haproxy metric types are sent to Kafka. At this moment, the list is hardcoded as a `wantedMetrics` slice:
    
    wantedMetrics  := []string{ "Scur", "Qcur","Smax","Slim","Weight","Qtime","Ctime","Rtime","Ttime","Req_rate","Req_rate_max","Req_tot","Rate","Rate_lim","Rate_max" }

For an explanation of the metric types, please read [this](http://cbonte.github.io/haproxy-dconv/configuration-1.5.html#9.1)            
### Updating the configuration via REST

Post a configuration. You can use the example file `resources/config_example.json`

    $ http POST http://192.168.59.103:10001/v1/config < resources/config_example.json 
    HTTP/1.1 200 OK
     
    Ok
    
Update the weight of some backend server

    $ http POST http://192.168.59.103:10001/v1/backend/testbe/servers/test_be_1/weight/20
    HTTP/1.1 200 OK
    
    Ok

### Running as local proxy + Zookeeper

At [magnetic.io](http://magnetic.io), we use Haproxy-rest running in local proxy mode for simple service discovery.
When you start HAproxy-rest with `-mode=localproxy`, only very simple binds are set up between two host:port pairs.
No frontends, no backends, no ACL's, no nothing.  

__Note:__ local proxy mode requires a Zookeeper ensemble to function: local proxy only gets its config from a Zookeeper
node.  

Haproxy-rest will watch for changes to the key: `/magnetic/localproxy`. You can set your own namespace using the `-zooConKey` flag.  The `/localproxy` part is hardcoded.
To this node you need to publish a full configuration in JSON format. Starting up a localproxy using Zookeeper
looks like this:  

    -mode=localproxy -zooConString=10.161.63.88:2181,10.189.106.106:2181,10.5.99.23:2181
    
This will result in config similar to the following JSON. Notice the `frontends` and `backends` are empty.
There is just a simple array of services that bind a port to an endpoint.

    {
        frontends: [ ],
        backends: [ ],
        services: [
            {
                name: "vrn-development-service-4d7a24cd",
                bindPort: 22500,
                endPoint: "10.224.236.38",
                mode: "tcp"
            }
        ]
    }
    
### Setting ACL's
    
You can set ACLs as part of a frontend's configuration and use these ACLs to route traffic to different backends.
The example below will route all Internet Explorer users to a different backend. You can update this on the fly
without loosing sessions or causing errors due to Haproxy's smart restart mechanisms.

    {
        "frontends" : [
            {
                "name" : "test_fe_1",                               # declare a frontend
                ...                                                 # some stuff left out for brevity
                "acls" : [
                    {
                        "name" : "uses_msie",                       # set an ACL by giving it a name and some pattern. 
                        "backend" : "testbe2",                      # set the backend to send traffic to
                        "pattern" : "hdr_sub(user-agent) MSIE"      # This pattern matches all HTTP requests that have
                    }                                               # "MSIE" in their User-Agent header                 

                ]
            }
        ]
    }

### Rate / Spike limiting 

You can set limits on specific connection rates for HTTP and TCP traffic. This comes in handy if you want to protect
yourself from abusive users or other spikes. The rates are calculated over a specific time range. The example below
tracks the TCP connection rate over 30 seconds. If more than 200 new connections are made in this time period, the 
client receives an 503 error and goes into a "cooldown" period for 60 seconds (`expiryTime`)

    {
        "frontends" : [
            {
                "name" : "test_fe_1",
                ... 
                "httpSpikeLimit" : {
                    "sampleTime" : "30s",
                    "expiryTime" : "60s",
                    "rate" : 50
                },
                "tcpSpikeLimit" : {
                    "sampleTime" : "30s",
                    "expiryTime" : "60s",
                    "rate" : 200
            }
    }


Note: the time format used, i.e. `30s`, is the default Haproxy time format. More details [here](http://cbonte.github.io/haproxy-dconv/configuration-1.5.html#2.2)
 
### Startup Flags & Options

    -binary="/usr/local/bin/haproxy"                           Path to the HAproxy binary
    -kafkaHost="localhost"                                     The hostname or ip address of the Kafka host
    -kafkaPort=9092                                            The port of the Kafka host
    -kafkaSwitch="off"                                         Switch whether to enable Kafka streaming
    -lbConfigFile="resources/haproxy_new.cfg"                  Location of the target HAproxy config file
    -lbTemplate="resources/haproxy_cfg.template"               Template file to build HAproxy load balancer config
    -mode="loadbalancer"                                       Switch for "loadbalancer" or "localproxy" mode
    -pidFile="resources/haproxy-private.pid"                   Location of the HAproxy PID file
    -port=10001                                                Port/IP to use for the REST interface. Overrides $PORT0 env variable
    -proxyConfigFile="resources/haproxy_localproxy_new.cfg"    Location of the target HAproxy localproxy config
    -proxyTemplate="resources/haproxy_localproxy_cfg.template" Template file to build HAproxy local proxy config
    -zooConKey="magnetic"                                      Zookeeper root key
    -zooConString="localhost"                                  A zookeeper ensemble connection string
    
for example, this would start up haproxy-rest on port 12345

    $ ./haproxy-rest -port=12345  
and this would start up haproxy-rest with kafka streaming enabled

    $ ./haproxy-rest -mode=loadbalancer -kafkaSwitch=on -kafkaHost=10.161.63.88
    
### Installing: the harder custom build way

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
    

## Inspiration

Part of Haproxy-rest is inspired by [haproxy-config](https://github.com/jbuchbinder/haproxy-config) and
[consul-haproxy](https://github.com/hashicorp/consul-haproxy). It is not a straight fork or clone of either of these,
but parts are borrowed.