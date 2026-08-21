// writetcp connectsto a TCP server at localhost with the specified port (8080 by default) and forwards stdin to the server
// line by line until EOF is reached.
// received lines from the server are printed to stdout
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	const name = "writetcp"
	log.SetPrefix(name + "\t")

	// register cli flag: -p for port
	port := flag.Int("p", 8080, "port to connect to")
	flag.Parse()

	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{Port: *port})
	if err != nil {
		log.Fatalf("error connecting to localhost:%d: %v", *port, err)
	}
	log.Printf("connected to %s: will forward stdin", conn.RemoteAddr())
	defer conn.Close()

	// scan a goroutine to read incoming lines
	// TCP is full-duplex, we can read and write simultaneously
	// we just need to spawn the gorouting to do the reading
	go func() {
		for connScanner := bufio.NewScanner(conn); connScanner.Scan(); {
			fmt.Printf("%s\n", connScanner.Text())

			if err := connScanner.Err(); err != nil {
				log.Fatalf("error reading from %s: %v", conn.RemoteAddr(), err)
			}
		}
	}()

	// read incoming lines from stdin and forward them to the server
	for stdinScanner := bufio.NewScanner(os.Stdin); stdinScanner.Scan(); {
		log.Printf("sent: %s\n", stdinScanner.Text())
		if _, err := conn.Write(stdinScanner.Bytes()); err != nil {
			log.Fatalf("error writing to %s: %v", conn.RemoteAddr(), err)
		}
		// add the new line back into the connection
		if _, err := conn.Write([]byte("\n")); err != nil {
			log.Fatalf("error writing to %s: %v", conn.RemoteAddr(), err)
		}
		if stdinScanner.Err() != nil {
			log.Fatalf("error reading from %s: %v", conn.RemoteAddr(), err)
		}
	}
}
