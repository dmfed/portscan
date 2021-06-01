package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/dmfed/portscan"
)

func main() {
	var (
		address = flag.String("addr", "127.0.0.1", "ip address to connect to")
		port    = flag.Int("p", 1, "port to connect to")
		endport = flag.Int("pe", 100, "ending port, if 0, then will only scan -p")
		maxconn = flag.Int("maxconn", 10, "maximum simultaneous connections to open")
		timeout = flag.Duration("t", time.Second, "connection timeout")
	)
	flag.Parse()
	_, err := net.ResolveIPAddr("ip", *address)
	if err != nil {
		fmt.Println("resolve err:", err)
		return
	}
	scanner := portscan.New(*address)
	if *endport < *port {
		*endport = *port
	}
	scanner.SetPorts(*port, *endport)
	scanner.SetMaxConn(*maxconn)
	scanner.SetTimeOut(*timeout)
	scanner.ScanAndPrint()
}
