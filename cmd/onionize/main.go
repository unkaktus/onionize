package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/nogoegst/fileserver"
	"github.com/nogoegst/onionize"
	"github.com/nogoegst/onionutil"
	"github.com/nogoegst/terminal"
	"github.com/nogoegst/textqr"
	"github.com/nogoegst/tlspin"
)

var debug bool

func main() {
	defaultLocalFlag := false
	defaultNoSlugFlag := false
	if strings.HasSuffix(os.Args[0], "expoze") {
		defaultLocalFlag = true
		defaultNoSlugFlag = true
	}
	var debugFlag = flag.Bool("debug", false,
		"Show what's happening")
	var noSlugFlag = flag.Bool("no-slug", defaultNoSlugFlag,
		"Do not use slugs")
	var zipFlag = flag.Bool("zip", false,
		"Serve zip file contents")
	var qrFlag = flag.Bool("qr", false,
		"Print link in QR code to stdout")
	var localFlag = flag.Bool("local", defaultLocalFlag,
		"Run in outside-reachable mode without onion service")
	var noTLSFlag = flag.Bool("no-tls", false,
		"Disable TLS")
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
		"Path to TLS private key")
	var tlspinKey = flag.String("tlspin-key", "",
		"tlspin private key (\"whateverkey\" to generate one)")
	var startTor = flag.Bool("start-tor", false,
		"start tor ourselves")
	flag.Parse()

	debug = *debugFlag
	paramsCh := make(chan onionize.Parameters)
	linkChan := make(chan url.URL)
	errChan := make(chan error)

	go func() {
		p := <-paramsCh
		go func() {
			errChan <- onionize.Onionize(p, linkChan)
		}()
	}()

	if len(flag.Args()) == 0 {
		guiMain(paramsCh, linkChan, errChan)
	} else {
		p := onionize.Parameters{
			Debug:           debug,
			ControlPath:     *control,
			ControlPassword: *controlPasswd,
			Pathspec:        fileserver.JoinPathspec(flag.Args()),
			Slug:            true,
			Zip:             *zipFlag,
			NoOnion:         *localFlag,
			StartTor:        *startTor,
		}
		if !(*noTLSFlag) { // TLS enabled
			// default to tlspin tofu by default for local mode
			if *tlspinKey == "" && *localFlag {
				*tlspinKey = "whatever"
			}

			if *tlsCertPath != "" && *tlsKeyPath != "" {
				var err error
				p.TLSConfig = &tls.Config{
					/*
						CipherSuites: []uint16{
							tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
							tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
						},
					*/
					Certificates: make([]tls.Certificate, 1),
				}
				p.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(*tlsCertPath, *tlsKeyPath)
				if err != nil {
					log.Fatalf("Unable to load TLS keypair: %v", err)
				}
			} else if *tlspinKey != "" {
				var err error
				p.TLSConfig, err = tlspin.TLSServerConfig(*tlspinKey)
				if err != nil {
					log.Fatalf("unable to load tlspin private key: %v", err)
				}
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

		if *noSlugFlag {
			p.Slug = false
		} else {
			// Disable slugs when using TLS
			if p.TLSConfig != nil {
				p.Slug = false
			}
		}

		paramsCh <- p

		for {
			select {
			case link := <-linkChan:
				linkString := link.String()
				if *qrFlag {
					textqr.Write(os.Stdout, linkString, textqr.L, true, false)
				}
				fmt.Println(linkString)

			case err := <-errChan:
				log.Fatal(err)
			}
		}
	}
}
