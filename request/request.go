package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	wg       sync.WaitGroup
	location string
)

const timeout = time.Duration(1 * time.Hour)

func init() {
	flag.StringVar(&location, "l", ".", "place to save")
}

var failed = 0

// A Buffer represents a target buffer
type Buffer struct {
	Body   io.Reader
	Dst    io.Writer
	Length int64
}

// Code from go/src/io/io.go
func copyBuffer(src Buffer) (written int64, err error) {
	size := 32 * 1024
	buf := make([]byte, size)
	if l, ok := src.Body.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	if buf == nil {
		buf = make([]byte, size)
	}
	fmt.Printf("\n")
	for {
		nr, er := src.Body.Read(buf)
		if nr > 0 {
			nw, ew := src.Dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
			fmt.Printf("progress: %d%%, %d bytes\r", written*100/src.Length, written)

		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

func request(uri string) {
	defer wg.Done()

	buf := Buffer{}
	url, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	tlsconfig := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := http.Client{
		Timeout:   timeout,
		Transport: tlsconfig,
	}

	res, err := cli.Get(url.String())
	buf.Body = res.Body

	len := res.Header.Get("Content-Length")
	if len != "" {
		len64, _ := strconv.ParseInt(len, 10, 64)
		buf.Length = len64
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\terr %s\n", r)
			failed++
		}
	}()
	defer res.Body.Close()

	index := strings.LastIndex(url.Path, "/")
	file := url.Path[index+1:]
	out, err := os.Create(location + "/" + file)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	buf.Dst = out

	var n int64
	if n, err = copyBuffer(buf); err != nil {
		panic(err)
	}
	fmt.Printf("+ %s %d bytes\n", uri[index+1:], n)
}

func main() {
	// Signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		select {
		case <-sig:
			time.Sleep(time.Millisecond * 1)
			os.Exit(1)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		wg.Add(1)
		go request(scanner.Text())
	}
	wg.Wait()
	fmt.Println("total failures:", failed)
}
