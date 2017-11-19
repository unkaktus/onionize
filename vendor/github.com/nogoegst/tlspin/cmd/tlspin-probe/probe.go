// probe.go - probe TLS public key in tlspin format.
//
// To the extent possible under law, Ivan Markin has waived all copyright
// and related or neighboring rights to tlspin, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	"flag"
	"log"

	"github.com/nogoegst/tlspin/util"
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatalf("address not specified")
	}
	addr := flag.Args()[0]
	conn, keydigest, err := tlspinutil.InitDial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()
	log.Printf("%s", tlspinutil.EncodeKey(keydigest))
}
