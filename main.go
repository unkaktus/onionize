package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/nogoegst/byteqr"
	libonionize "github.com/nogoegst/onionize/lib"
	"github.com/nogoegst/onionutil"
	"github.com/nogoegst/terminal"
	"rsc.io/qr"
)

var debug bool

func main() {
	var debugFlag = flag.Bool("debug", false,
		"Show what's happening")
	var noslugFlag = flag.Bool("noslug", false,
		"Do not use slugs")
	var zipFlag = flag.Bool("zip", false,
		"Serve zip file contents")
	var qrFlag = flag.Bool("qr", false,
		"Print link in QR code to stdout")
	var noOnionFlag = flag.Bool("noonion", false,
		"Run in outside-reachable mode without onion service")
	var passphraseFlag = flag.Bool("p", false,
		"Ask for passphrase to generate onion key")
	var control = flag.String("control-addr", "default://",
		"Set Tor control address to be used")
	var controlPasswd = flag.String("control-passwd", "",
		"Set Tor control auth password")
	var idKeyPath = flag.String("id-key", "",
		"Path to onion identity private key")
	var tlsCertPath = flag.String("tls-cert", "",
		"Path to TLS certificate")
	var tlsKeyPath = flag.String("tls-key", "",
		"Path tp TLS private key")
	flag.Parse()

	debug = *debugFlag
	paramsCh := make(chan libonionize.Parameters)
	linkChan := make(chan url.URL)
	errChan := make(chan error)

	go func() {
		p := <-paramsCh
		go func() {
			errChan <- libonionize.Onionize(p, linkChan)
		}()
	}()

	if len(flag.Args()) == 0 {
		guiMain(paramsCh, linkChan, errChan)
	} else {
		if len(flag.Args()) != 1 {
			log.Fatalf("You should specify exactly one path/target URL")
		}
		p := libonionize.Parameters{
			Debug:           debug,
			ControlPath:     *control,
			ControlPassword: *controlPasswd,
			Path:            flag.Args()[0],
			Slug:            !*noslugFlag,
			Zip:             *zipFlag,
			NoOnion:         *noOnionFlag,
		}
		if *tlsCertPath != "" && *tlsKeyPath != "" {
			var err error
			p.TLSConfig = &tls.Config{
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				},
				Certificates: make([]tls.Certificate, 1),
			}
			p.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(*tlsCertPath, *tlsKeyPath)
			if err != nil {
				log.Fatalf("Unable to load TLS keypair: %v", err)
			}
		}
		if *passphraseFlag {
			fmt.Fprintf(os.Stderr, "Enter your passphrase for onion identity: ")
			onionPassphrase, err := terminal.ReadPassword(0)
			if err != nil {
				log.Fatalf("Unable to read onion passphrase: %v", err)
			}
			fmt.Printf("\n")
			p.Passphrase = string(onionPassphrase)
		} else if *idKeyPath != "" {
			var err error
			p.IdentityKey, _, err = onionutil.LoadPrivateKeyFile(*idKeyPath)
			if err != nil {
				log.Fatalf("Unable to load identity private key: %v", err)
			}
		}

		paramsCh <- p

		for {
			select {
			case link := <-linkChan:
				linkString := link.String()
				if *qrFlag {
					byteqr.Write(os.Stdout, linkString, qr.L, nil, nil)
				}
				fmt.Println(linkString)

			case err := <-errChan:
				log.Fatal(err)
			}
		}
	}
}
