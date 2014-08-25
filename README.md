# HAproxy-rest
---

HAproxy-rest is a REST interface for HAproxy. It exports basic functions normally done through the config file or the
HAproxy's socket interface to a handy REST interface.

## Getting started

Startup Haproxy with the options to gracefull restart

OSX:

    $ haproxy -f resources/haproxy_init.cfg -p resources/haproxy-private.pid -st $(<resources/haproxy-private.pid)
Ubuntu
        
    $ haproxy -f resources/haproxy_init.cfg -p resources/haproxy-private.pid -sf $(cat resources/haproxy-private.pid)
    







## Inspiration

Part of Haproxy-rest is inspired by [https://github.com/jbuchbinder/haproxy-config](https://github.com/jbuchbinder/haproxy-config).
It is not a straight fork, but the Config and Persistence parts are reused.