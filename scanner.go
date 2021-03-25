package main

import (
	"flag"
	"fmt"
	"net"
	"time"
)

func connect(addr net.TCPAddr, timeout time.Duration, result chan string, done chan bool) {
	conn, err := net.DialTimeout("tcp", addr.String(), timeout)
	if err == nil {
		result <- fmt.Sprintf("%v accepting connections", addr.String())
		conn.Close()
	} /* else {
		result <- "" //err.Error()
	}*/
	done <- true
}

func scanPorts(addr *net.IPAddr, start, end, maxconn int, connect func(net.TCPAddr, chan string, chan bool)) chan string {
	comm := make(chan string, maxconn)
	done := make(chan bool, maxconn)
	open := 0
	go func() {
		for start <= end {
			if open < maxconn {
				tcp := net.TCPAddr{IP: addr.IP, Port: start}
				go connect(tcp, comm, done)
				start++
				open++
			}
			select {
			case <-done:
				open--
			default:
			}
		}
		for i := 0; i < open; i++ {
			<-done // Make sure all connecting goroutines finished
		}
		close(done)
		close(comm)
	}()
	return comm
}

func getConnFunc(timeout time.Duration) func(net.TCPAddr, chan string, chan bool) {
	return func(addr net.TCPAddr, result chan string, done chan bool) {
		connect(addr, timeout, result, done)
	}
}

func getScanner(start, end, maxconn int, connect func(net.TCPAddr, chan string, chan bool)) func(*net.IPAddr) chan string {
	return func(ip *net.IPAddr) chan string {
		return scanPorts(ip, start, end, maxconn, connect)
	}
}

func readAndPrint(input chan string) {
	t := time.Now()
	for result := range input {
		fmt.Println(result)
	}
	fmt.Println("scanned in:", time.Since(t))
}

func main() {
	var (
		address = flag.String("addr", "127.0.0.1", "ip address to connect to")
		port    = flag.Int("p", 22, "port to connect to")
		endport = flag.Int("pe", 0, "ending port, if 0, then will only scan -port")
		maxconn = flag.Int("maxconn", 10, "maximum simultaneous connections to open")
		timeout = flag.Duration("t", time.Second, "connection timeout")
	)
	flag.Parse()
	ip, err := net.ResolveIPAddr("ip", *address)
	if err != nil {
		fmt.Println("resolve err:", err)
		return
	}
	if *endport < *port {
		*endport = *port
	}
	connfunc := getConnFunc(*timeout)
	scanner := getScanner(*port, *endport, *maxconn, connfunc)
	readAndPrint(scanner(ip))
}
