// onionize.go - onionize directories.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	"log"
	"fmt"
	"flag"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"os"
	"path/filepath"
	"archive/zip"
	"crypto/rand"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/tools/godoc/vfs"
	"github.com/nogoegst/pickfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"github.com/nogoegst/onionutil"
	"github.com/nogoegst/bulb"
	bulbUtils "github.com/nogoegst/bulb/utils"
)

func main () {
	var debugFlag = flag.Bool("debug", false,
		"Show what's happening")
	var zipFlag = flag.Bool("zip", false,
		"Serve zip file contents")
	var passphraseFlag = flag.Bool("p", false,
		"Ask for passphrase to generate onion key")
	var control = flag.String("control-addr", "tcp://127.0.0.1:9051",
		"Set Tor control address to be used")
	var controlPasswd = flag.String("control-passwd", "",
		"Set Tor control auth password")
	flag.Parse()
	if (len(flag.Args()) != 1) {
		log.Fatalf("You should specify exacly one webroot path")
	}
	pathToServe := flag.Args()[0]
	debug := *debugFlag
	// Parse control string
	controlNet, controlAddr, err := bulbUtils.ParseControlPortString(*control)
	if err != nil {
	log.Fatalf("Failed to parse Tor control address string: %v", err)
	}
	// Connect to a running tor instance.
	c, err := bulb.Dial(controlNet, controlAddr)
	if err != nil {
		log.Fatalf("Failed to connect to control socket: %v", err)
	}
	defer c.Close()

	// See what's really going on under the hood.
	// Do not enable in production.
	c.Debug(debug)

	// Authenticate with the control port.  The password argument
	// here can be "" if no password is set (CookieAuth, no auth).
	if err := c.Authenticate(*controlPasswd); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// At this point, c.Request() can be used to issue requests.

	var fs vfs.FileSystem
	var url string

	if *zipFlag {
		rcZip, err := zip.OpenReader(pathToServe)
		if err != nil {
			log.Fatalf("Unable to open zip archive: %v", err)
		}
		fs = zipfs.New(rcZip, "onionize")
	} else {
		fileInfo, err := os.Stat(pathToServe)
		if err != nil {
			log.Fatalf("Unable to open path: %v", err)
		}
		if fileInfo.IsDir() {
			fs = vfs.OS(pathToServe)
		} else {
			abspath, err := filepath.Abs(pathToServe)
			if err != nil {
				log.Fatalf("Unable to get absolute path to file")
			}
			dir, file := filepath.Split(abspath)
			slugBin := make([]byte, 5)
			_, err = rand.Read(slugBin)
			slug := onionutil.Base32Encode(slugBin)[:8]
			m := make(map[string]string)
			url = slug+"/"+file
			m[url]= file
			fs = pickfs.New(vfs.OS(dir), m)
			escapedFilename := strings.Replace(neturl.QueryEscape(file), "+", "%20", -1)
			url = slug+"/"+escapedFilename
		}
	}
	http.Handle("/", http.FileServer(httpfs.New(fs)))

	var onionListener net.Listener

	if *passphraseFlag {
		fmt.Fprintf(os.Stderr, "Enter your passphrase for onion identity: ")
		onionPassphrase, err := terminal.ReadPassword(0)
		if err != nil {
			log.Fatalf("Unable to read onion passphrase: %v", err)
		}
		fmt.Printf("\n")

		privOnionKey, err := onionutil.GenerateOnionKey(onionutil.KeystreamReader([]byte(onionPassphrase), []byte("onionize-keygen")))
		if err != nil {
			log.Fatalf("Unable to generate onion key: %v", err)
		}
		onionListener, err = c.Listener(80, privOnionKey)
	} else {
		onionListener, err = c.Listener(80, nil)
	}
        if err != nil {
                log.Fatalf("Error occured while creating an onion service: %v", err)
        }
        defer onionListener.Close()
	onionID, _, err := net.SplitHostPort(onionListener.Addr().String())

	c.StartAsyncReader()
        if err != nil {
		log.Fatalf("Unable to derive onionID from listener.Addr(): %v", err)
        }
	if _, err := c.Request("SETEVENTS HS_DESC"); err != nil {
		log.Fatalf("SETEVENTS HS_DESC has failed: %v", err)
	}

	for {
		ev, err := c.NextEvent()
		if err != nil {
			log.Printf("NextEvent has failed: %v", err)
			continue
		}
		splittedReply := strings.Split(ev.Reply, " ")
		if (len(splittedReply) < 3) {
			continue
		}
		hsDescAction := splittedReply[1]
		if (hsDescAction != "UPLOADED") {
			continue
		}
		onionID := splittedReply[2]
		if (onionID != onionListener.Addr().String()[:len(onionID)]) {
			continue
		}
		break
	}
        fmt.Printf("http://%s/%s\n", onionID, url)

        err = http.Serve(onionListener, nil)
        if err != nil {
                log.Fatalf("Cannot serve HTTP")
        }

}

