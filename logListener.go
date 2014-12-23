package main

import (
	"fmt"
	"net"
)

func logPrinter(c *net.UnixConn) {
	for {
		var buf [1024]byte
		nr, err := c.Read(buf[:])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", string(buf[:nr]))
	}
}

func logListener(socket string) {
	log.Info("Starting log listener")

	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{socket, "unixgram"})
	if err != nil {
		log.Fatal("listen error:", err)
	}

	go logPrinter(conn)

}

