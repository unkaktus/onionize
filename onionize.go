// onionize.go - onionize things.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionize

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/nogoegst/fileserver"
	"github.com/nogoegst/onionutil"
	"github.com/yawning/bulb"
)

const slugLength = 16

type Parameters struct {
	Pathspec        string
	Zip             bool
	Slug            bool
	ControlPath     string
	ControlPassword string
	Passphrase      string
	Debug           bool
	IdentityKey     crypto.PrivateKey
	TLSConfig       *tls.Config
	NoOnion         bool
}

func generateSlug() (string, error) {
	slugBin := make([]byte, (slugLength*5)/8+1)
	_, err := rand.Read(slugBin)
	if err != nil {
		return "", err
	}
	return onionutil.Base32Encode(slugBin)[:slugLength], nil
}

func Onionize(p Parameters, linkChan chan<- url.URL) error {
	var handler http.Handler
	var slug string
	var err error
	if p.Slug && !p.NoOnion {
		slug, err = generateSlug()
		if err != nil {
			return fmt.Errorf("Unable to generate slug: %v", err)
		}
	}

	useOnion := !p.NoOnion
	link := url.URL{Path: "/"}
	var c *bulb.Conn
	nocfg := &bulb.NewOnionConfig{
		DiscardPK:      true,
		AwaitForUpload: true,
	}

	if strings.HasPrefix(p.Pathspec, "http://") || strings.HasPrefix(p.Pathspec, "https://") {
		target, err := url.Parse(p.Pathspec)
		if err != nil {
			return fmt.Errorf("Unable to parse target URL: %v", err)
		}
		handler = onionReverseHTTPProxy(target)
	} else {
		handler, err = fileserver.New(p.Pathspec, p.Zip, p.Debug)
		if err != nil {
			return err
		}
	}
	server := &http.Server{Handler: subdomainSluggedHandler(handler, slug)}

	listenAddress := "127.0.0.1:0"
	if useOnion {
		// Connect to a running tor instance
		if p.ControlPath == "" {
			p.ControlPath = "default://"
		}
		c, err = bulb.DialURL(p.ControlPath)
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
		if p.Passphrase != "" {
			keyrd, err := onionutil.KeystreamReader([]byte(p.Passphrase), []byte("onionize-keygen"))
			if err != nil {
				return fmt.Errorf("Unable to create keystream: %v", err)
			}
			privOnionKey, err := onionutil.GenerateOnionKey(keyrd, "current")
			if err != nil {
				return fmt.Errorf("Unable to generate onion key: %v", err)
			}
			nocfg.PrivateKey = privOnionKey
		} else {
			nocfg.PrivateKey = p.IdentityKey
		}
	} else {
		tc, err := net.Dial("udp", "1.1.1.1:1")
		if err != nil {
			return err
		}
		defer tc.Close()
		host, _, err := net.SplitHostPort(tc.LocalAddr().String())
		if err != nil {
			return err
		}
		listenAddress = host + ":0"
	}

	var listener net.Listener
	rawListener, err := net.Listen("tcp4", listenAddress)
	if err != nil {
		return err
	}

	var virtPort uint16
	if p.TLSConfig != nil {
		listener = tls.NewListener(rawListener, p.TLSConfig)
		link.Scheme = "https"
		virtPort = uint16(443)
	} else {
		listener = rawListener
		link.Scheme = "http"
		virtPort = uint16(80)
	}

	if useOnion {
		portSpec := bulb.OnionPortSpec{
			VirtPort: virtPort,
			Target:   listener.Addr().String(),
		}
		nocfg.PortSpecs = []bulb.OnionPortSpec{portSpec}
		oi, err := c.NewOnion(nocfg)
		if err != nil {
			return fmt.Errorf("Error occured while creating an onion service: %v", err)
		}
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

		if slug != "" {
			link.Host = fmt.Sprintf("%s.%s.onion", slug, oi.OnionID)
		} else {
			link.Host = fmt.Sprintf("%s.onion", oi.OnionID)
		}
	} else {
		link.Host = listener.Addr().String()
	}

	// Return the link to the service
	linkChan <- link
	// Run a webservice
	err = server.Serve(listener)
	if err != nil {
		return fmt.Errorf("Cannot serve HTTP: %v", err)
	}
	return nil
}
