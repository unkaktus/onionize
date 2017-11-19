// qr.go - make an QR from arguments
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to byteqr, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/nogoegst/byteqr"
	"rsc.io/qr"
)

func main() {
	flag.Parse()
	data := strings.Join(flag.Args(), " ")

	err := byteqr.Write(os.Stdout, data, qr.L, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
}
