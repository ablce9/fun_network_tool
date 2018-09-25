// Simple proxy server
// stallion.local has two net interfaces 10.42.0.1 and 192.168.0.1 but poney.local doesn't have
// access to 192.168.0.1 for some reason. Then, stallion.local can run me:
// `go run proxy.go -address 192.168.0.1 -dst 10.42.0.1`
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
)

var (
	srcAddress string
	srcPort    string
	dstAddress string
	dstPort    string
)

func init() {
	flag.StringVar(&srcAddress, "address", "0.0.0.0", "bind to this address")
	flag.StringVar(&srcPort, "port", "9000", "bind to this port")
	flag.StringVar(&dstAddress, "dst", "0.0.0.0", "destination address")
	flag.StringVar(&dstPort, "dstport", "80", "destination port")
}

func main() {
	flag.Parse()

	fmt.Printf("server address=%s:%s target=%s:%s\n", srcAddress, srcPort, dstAddress, dstPort)

	ln, err := net.Listen("tcp", net.JoinHostPort(srcAddress, srcPort))
	defer ln.Close()

	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			panic(err)
		}
		go handleListener(conn, net.JoinHostPort(dstAddress, dstPort))
	}
}

func handleListener(src net.Conn, addr string) error {
	fmt.Printf("getting request from %s\n", src.RemoteAddr())
	errCh := make(chan error, 2)
	dst, err := net.Dial("tcp", addr)
	defer dst.Close()

	if err != nil {
		panic(err)
	}

	go doProxy(dst, src, errCh)
	go doProxy(src, dst, errCh)

	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			return e
		}
	}
	return nil
}

type closeWriter interface {
	CloseWrite() error
}

func doProxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)

	if conn, ok := dst.(closeWriter); ok {
		conn.CloseWrite()
	}

	errCh <- err
}
