package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	libonionize "github.com/nogoegst/onionize/lib"
	"github.com/nogoegst/terminal"
)

var debug bool

func main() {
	var debugFlag = flag.Bool("debug", false,
		"Show what's happening")
	var noslugFlag = flag.Bool("noslug", false,
		"Do not use slugs")
	var zipFlag = flag.Bool("zip", false,
		"Serve zip file contents")
	var passphraseFlag = flag.Bool("p", false,
		"Ask for passphrase to generate onion key")
	var control = flag.String("control-addr", "default://",
		"Set Tor control address to be used")
	var controlPasswd = flag.String("control-passwd", "",
		"Set Tor control auth password")
	flag.Parse()

	debug = *debugFlag
	paramsCh := make(chan libonionize.Parameters)
	linkChan := make(chan string)
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
		}
		if *passphraseFlag {
			fmt.Fprintf(os.Stderr, "Enter your passphrase for onion identity: ")
			onionPassphrase, err := terminal.ReadPassword(0)
			if err != nil {
				log.Fatalf("Unable to read onion passphrase: %v", err)
			}
			fmt.Printf("\n")
			p.Passphrase = string(onionPassphrase)
		}
		paramsCh <- p

		for {
			select {
			case link := <-linkChan:
				fmt.Println(link)
			case err := <-errChan:
				log.Fatal(err)
			}
		}
	}
}
