package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	const name = "tcpupperecho"
	log.SetPrefix(name + "\t")

	port := flag.Int("p", 8080, "port to listen on")
	flag.Parse()
	// ListenTCP creates a TCP listener accepting connections on given address
	// TCPAddr - address of tcp endpoing with IP, Port, and Zone all options
	// zone is for ipv6
	// if we omit ip, we listen on all available ip addresses
	// if we omit port, we listen on a random port
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: *port})
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	log.Printf("listening at localhost:%s", listener.Addr())
	for {
		// loop forever accepting connections one at a time
		// Accept() blocks until connection is made then reterns the conn
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go echoUpper(conn, conn)
	}
}

func echoUpper(w io.Writer, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// scanner.Text() strips the newline char from the end of the line
		// so we need to add it back in when we write to w
		// writing to w does not print to the server terminal
		// it writes to the connection (responding to the client)
		fmt.Fprintf(w, "%s\n", strings.ToUpper(line))
		// log the output to the server 
		log.Printf("recieved: %s", line)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error: %s", err)
	}
}
