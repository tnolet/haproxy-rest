package main

import (
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}


type ZookeeperClient struct {

	conString string

}

func (z ZookeeperClient) connect() *zk.Conn {
	zks := strings.Split(z.conString, ",")
	conn, _, err := zk.Connect(zks, (60 * time.Second))
	must(err)
	return conn
}

/**
 * Watches a Zookeeper node continuously in a loop. When a watch fires, the new config is rendered.
 * When first registering the watch, the initial payload is also rendered
 */
func (z ZookeeperClient) watchLocalProxyConfig(conn *zk.Conn, path string) {

	go func() {
		for {
			payload, _, watch, err := conn.GetW(path)
			must(err)

			RenderLocalProxyConfig(payload, ConfigObj)
			// block till event fires
			event := <- watch

			log.Info("Received Zookeeper event: " + event.Type.String())
			RenderLocalProxyConfig(payload,ConfigObj)
		}
	}()

}


