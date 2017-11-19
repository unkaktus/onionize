package main

import (
	"flag"
	"log"

	"github.com/nogoegst/tlspin"
)

func main() {
	flag.Parse()
	sk := flag.Args()[0]
	l, err := tlspin.Listen("tcp", ":8443", sk)
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			c.Write([]byte("hello"))
		}()
	}
}
