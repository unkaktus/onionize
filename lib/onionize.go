// onionize.go - onionize directories, files and zips.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionize

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/nogoegst/bulb"
	"github.com/nogoegst/onionutil"
)

const slugLengthB32 = 16

type Parameters struct {
	Path            string
	Zip             bool
	Slug            bool
	ControlPath     string
	ControlPassword string
	Passphrase      string
	Debug           bool
}

type ResultLink struct {
	URL   string
	Error error
}

func Onionize(p Parameters, linkCh chan<- ResultLink) {
	var handler http.Handler
	var link string
	target, err := url.Parse(p.Path)
	if err != nil {
		linkCh <- ResultLink{Error: fmt.Errorf("Unable to parse target URL: %v", err)}
		return
	}
	switch target.Scheme {
	case "http", "https":
		handler = OnionReverseHTTPProxy(target)
	default:
		handler, link, err = FileServer(p.Path, p.Slug, p.Zip, p.Debug)
		if err != nil {
			linkCh <- ResultLink{Error: err}
			return
		}
	}
	server := &http.Server{Handler: handler}

	// Connect to a running tor instance
	if p.ControlPath == "" {
		p.ControlPath = "default://"
	}
	c, err := bulb.DialURL(p.ControlPath)
	if err != nil {
		linkCh <- ResultLink{Error: fmt.Errorf("Failed to connect to control socket: %v", err)}
		return
	}
	defer c.Close()

	// See what's really going on under the hood
	c.Debug(p.Debug)

	// Authenticate with the control port
	if err := c.Authenticate(p.ControlPassword); err != nil {
		linkCh <- ResultLink{Error: fmt.Errorf("Authentication failed: %v", err)}
		return
	}
	// Derive onion service keymaterial from passphrase or generate a new one
	aocfg := &bulb.NewOnionConfig{
		DiscardPK:      true,
		AwaitForUpload: true,
	}
	if p.Passphrase != "" {
		keyrd, err := onionutil.KeystreamReader([]byte(p.Passphrase), []byte("onionize-keygen"))
		if err != nil {
			linkCh <- ResultLink{Error: fmt.Errorf("Unable to create keystream: %v", err)}
			return
		}
		privOnionKey, err := onionutil.GenerateOnionKey(keyrd, "current")
		if err != nil {
			linkCh <- ResultLink{Error: fmt.Errorf("Unable to generate onion key: %v", err)}
			return
		}
		aocfg.PrivateKey = privOnionKey
	}
	onionListener, err := c.NewListener(aocfg, 80)
	if err != nil {
		linkCh <- ResultLink{Error: fmt.Errorf("Error occured while creating an onion service: %v", err)}
		return
	}
	defer onionListener.Close()
	// Track if tor went down
	go func() {
		for {
			_, err := c.NextEvent()
			if err != nil {
				log.Fatalf("Lost connection to tor: %v", err)
			}
		}
	}()
	onionHost := strings.TrimSuffix(onionListener.Addr().String(), ":80")

	// Return the link to the service
	linkCh <- ResultLink{URL: fmt.Sprintf("http://%s/%s", onionHost, link)}
	// Run a webservice
	err = server.Serve(onionListener)
	if err != nil {
		log.Fatalf("Cannot serve HTTP")
	}
}
