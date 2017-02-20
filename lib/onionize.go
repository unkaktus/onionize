// onionize.go - onionize things.
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

func Onionize(p Parameters, linkChan chan<- string) error {
	var handler http.Handler
	var link string
	target, err := url.Parse(p.Path)
	if err != nil {
		return fmt.Errorf("Unable to parse target URL: %v", err)
	}
	switch target.Scheme {
	case "http", "https":
		handler = OnionReverseHTTPProxy(target)
	default:
		handler, link, err = FileServer(p.Path, p.Slug, p.Zip, p.Debug)
		if err != nil {
			return err
		}
	}
	server := &http.Server{Handler: handler}

	// Connect to a running tor instance
	if p.ControlPath == "" {
		p.ControlPath = "default://"
	}
	c, err := bulb.DialURL(p.ControlPath)
	if err != nil {
		return fmt.Errorf("Failed to connect to control socket: %v", err)
	}
	defer c.Close()

	// See what's really going on under the hood
	c.Debug(p.Debug)

	// Authenticate with the control port
	if err := c.Authenticate(p.ControlPassword); err != nil {
		return fmt.Errorf("Authentication failed: %v", err)
	}
	// Derive onion service keymaterial from passphrase or generate a new one
	aocfg := &bulb.NewOnionConfig{
		DiscardPK:      true,
		AwaitForUpload: true,
	}
	if p.Passphrase != "" {
		keyrd, err := onionutil.KeystreamReader([]byte(p.Passphrase), []byte("onionize-keygen"))
		if err != nil {
			return fmt.Errorf("Unable to create keystream: %v", err)
		}
		privOnionKey, err := onionutil.GenerateOnionKey(keyrd, "current")
		if err != nil {
			return fmt.Errorf("Unable to generate onion key: %v", err)
		}
		aocfg.PrivateKey = privOnionKey
	}
	onionListener, err := c.NewListener(aocfg, 80)
	if err != nil {
		return fmt.Errorf("Error occured while creating an onion service: %v", err)
	}
	defer onionListener.Close()
	// Track if tor went down
	// TODO: Signal from here to perform graceful shutdown and display a message
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
	linkChan <- fmt.Sprintf("http://%s/%s", onionHost, link)
	// Run a webservice
	err = server.Serve(onionListener)
	if err != nil {
		return fmt.Errorf("Cannot serve HTTP: %v", err)
	}
	return nil
}
