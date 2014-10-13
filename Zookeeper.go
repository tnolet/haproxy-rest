package main

import (
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
	"encoding/json"
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

func (z ZookeeperClient) watchServiceTypes(conn *zk.Conn, path string)(chan []string, chan error) {

	snapshots := make(chan []string)
	errors := make(chan error)
	go func() {
		for {
			snapshot, _, events, err := conn.ChildrenW(path)
			if err != nil {
				errors <- err
				return
			}
			snapshots <- snapshot
			evt := <-events
			log.Info("Received Zookeeper event: " + evt.Type.String())
			if evt.Err != nil {
				errors <- evt.Err
				return
			}
		}
	}()
	return snapshots, errors

}
func (z ZookeeperClient) watchHandler(conn *zk.Conn, snapshots chan []string, errors chan error) {


	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				log.Debug("Found services %+v\n", snapshot)

				// create list of known services
				var knownServices []string
				for _, knownService := range ConfigObj.Services {
					knownServices = AppendString(knownServices, knownService.Name)
				}

				// check if there are new services in Zookeeper compared to our local config,
				// if so, add it to the config
				for _, service := range snapshot {

					if stringInSlice(service, knownServices) {

						log.Debug("Already know service: " + service)

					} else {
						result,_, err := conn.Get("/magnetic/" + service)
						must(err)

						var s Service
						json.Unmarshal(result, &s)

						err = AddServiceToConfig(s.Name, s.BindPort, s.EndPoint, s.Mode, ConfigObj)
						must(err)
					}
				}

				// check if there are service in our local config no longer in Zookeeper,
				// if so, remove them from the config

				for _, knownService := range knownServices {

					if stringInSlice(knownService, snapshot) {

						log.Info("Services are matched, no removal triggered")

					} else {

						log.Info("Service " + knownService + "is not in Zookeeper anymore. Removing from proxy config.")
						RemoveServiceFromConfig(knownService, ConfigObj)
					}
				}

			case err := <-errors:
				log.Error("Caught errror from Zookeeper" + err.Error())

				// if sesssion is expired, reconnect.
				if (err == zk.ErrSessionExpired) {
					log.Error("Session was expired. Should restart")
					conn.Close()

					zkConnection := zkClient.connect()
					defer zkConnection.Close()

					snapshots, errors := zkClient.watchServiceTypes(zkConnection,"/magnetic")
					zkClient.watchHandler(zkConnection, snapshots, errors)

				}
			}
		}
	}()

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func AppendString(slice []string, elements ...string) []string {
	n := len(slice)
	total := len(slice) + len(elements)
	if total > cap(slice) {
		// Reallocate. Grow to 1.5 times the new size, so we can still grow.
		newSize := total*3/2 + 1
		newSlice := make([]string, total, newSize)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[:total]
	copy(slice[n:], elements)
	return slice
}

