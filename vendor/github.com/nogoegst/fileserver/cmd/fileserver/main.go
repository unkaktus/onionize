package main

import (
	"flag"
	"log"
	"net"

	"github.com/nogoegst/fileserver"
)

func main() {
	var zipFlag = flag.Bool("z", false, "serve from zip archive")
	var debugFlag = flag.Bool("debug", false, "debug")
	flag.Parse()
	path := fileserver.JoinPathspec(flag.Args())
	l, err := net.Listen("tcp4", "127.0.0.1:9999")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Serving on %s", l.Addr().String())
	log.Fatal(fileserver.Serve(l, path, *zipFlag, *debugFlag))
}
