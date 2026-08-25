package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	host, path, method string
	port               int
)

func main() {
	flag.StringVar(&method, "method", "GET", "HTTP method to use")
	flag.StringVar(&host, "host", "localhost", "host to connect to")
	flag.IntVar(&port, "port", 8080, "port to connect to")
	flag.StringVar(&path, "path", "/", "path to request")
	flag.Parse()

	ip, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, ip)
	if err != nil {
		panic(err)
	}

	log.Printf("connected to %s (@ %s)", host, conn.RemoteAddr())

	defer conn.Close()

	var reqfields = []string{
		fmt.Sprintf("%s %s HTTP/1.1", method, path),
		"HOST: " + host,
		"User-Agent: httpget",
		"", // empty line terminates the headers
		// body would go here if we had one
	}

	request := strings.Join(reqfields, "\r\n") + "\r\n"

	conn.Write([]byte(request))
	log.Printf("sent request:\n%s", request)

	for scanner := bufio.NewScanner(conn); scanner.Scan(); {
		line := scanner.Bytes()
		if _, err := fmt.Fprintf(os.Stdout, "%s\n", line); err != nil {
			log.Printf("error writing to connectin: %s", err)
		}
		if scanner.Err() != nil {
			log.Printf("error reading from connection: %s", err)
			return
		}
	}
}
