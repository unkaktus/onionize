package main

import (
	"flag"
	"log"

	"github.com/nogoegst/tlspin"
)

func main() {
	flag.Parse()
	pk := flag.Args()[0]
	conn, err := tlspin.Dial("tcp", "localhost:8443", pk)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	buf := make([]byte, 255)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", buf[:n])
}
