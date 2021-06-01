package portscan

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Scanner struct {
	IP        net.IP
	StartPort int
	EndPort   int
	Maxconn   int
	Timeout   time.Duration
}

func New(ip ...string) *Scanner {
	var s Scanner
	if len(ip) > 0 {
		s.IP = net.ParseIP(ip[0])
	} else {
		s.IP = net.ParseIP("127.0.0.1")
	}
	s.StartPort = 0
	s.EndPort = 1000
	s.Maxconn = 1
	s.Timeout = time.Second
	return &s
}

// Scan returns slice of net.TCPAddr which accepted connection
func (s *Scanner) Scan() []net.TCPAddr {
	addrchan := scanPorts(s.IP, s.StartPort, s.EndPort, s.Maxconn, s.Timeout)
	acceptingAddresses := []net.TCPAddr{}
	for addr := range addrchan {
		acceptingAddresses = append(acceptingAddresses, addr)
	}
	return acceptingAddresses
}

// ScanAndPrint scans addresses and instantly prints out the results
// Having finished it prints out total time taken by test
func (s *Scanner) ScanAndPrint() {
	fmt.Println("Scanning", s.IP)
	t := time.Now()
	addrchan := scanPorts(s.IP, s.StartPort, s.EndPort, s.Maxconn, s.Timeout)
	for addr := range addrchan {
		fmt.Printf("%s accepting connection\n", addr.String())
	}
	fmt.Println("scanned in:", time.Since(t))
}

func (s *Scanner) SetIP(ip string) {
	s.IP = net.ParseIP(ip)
}

func (s *Scanner) SetPorts(start, end int) {
	s.StartPort = start
	s.EndPort = end
}

func (s *Scanner) SetMaxConn(maxconn int) {
	s.Maxconn = maxconn
}

func (s *Scanner) SetTimeOut(t time.Duration) {
	s.Timeout = t
}

func scanPorts(addr net.IP, start, end, maxconn int, timeout time.Duration) chan net.TCPAddr {
	comm := make(chan net.TCPAddr, maxconn)
	go func() {
		open := make(chan struct{}, maxconn) // semaphore
		var wg sync.WaitGroup
		for start <= end {
			open <- struct{}{} // acquire
			tcp := net.TCPAddr{IP: addr, Port: start}
			wg.Add(1)
			start++
			go func() {
				conn, err := net.DialTimeout("tcp", tcp.String(), timeout)
				if err == nil {
					comm <- tcp
					conn.Close()
				}
				wg.Done()
				<-open // release
			}()
		}
		wg.Wait()
		close(comm)
	}()
	return comm
}
