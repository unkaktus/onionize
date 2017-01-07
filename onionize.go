// onionize.go - onionize directories, files and zips.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package main

import (
	"archive/zip"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"crypto/subtle"

	"github.com/nogoegst/bulb"
	"github.com/nogoegst/onionutil"
	"github.com/nogoegst/pickfs"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

const slugLengthB32 = 16

type Parameters struct {
	Path	string
	Zip	bool
	Slug	bool
	ControlPath	string
	ControlPassword	string
	Passphrase	string
}


func ResetHTTPConn(w *http.ResponseWriter) error {
	hj, ok := (*w).(http.Hijacker)
	if !ok {
		return fmt.Errorf("This webserver doesn't support hijacking")
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func CheckAndRewriteSlug(req *http.Request, slug string) error {
	if slug == "" {
		return nil
	}
	reqURL := strings.TrimLeft(req.URL.String(), "/")
	if len(reqURL) < len(slug) {
		return fmt.Errorf("URL is too short to have a slug in it")
	}
	if 1 != subtle.ConstantTimeCompare([]byte(slug), []byte(reqURL[:len(slug)])) {
		return fmt.Errorf("Wrong slug")
	}
	reqURL = strings.TrimPrefix(reqURL, slug)
	req.URL, _ = neturl.Parse(reqURL)
	return nil
}

type ResultLink struct {
	URL	string
	Error	error
}

func Onionize(p Parameters, linkCh chan<- ResultLink) {
	var fs vfs.FileSystem
	var url string
	var slug string
	if p.Slug {
		slugBin := make([]byte, (slugLengthB32*5)/8+1)
		_, err := rand.Read(slugBin)
		if err != nil {
			log.Fatalf("Unable to generate slug: %v", err)
		}
		slug = onionutil.Base32Encode(slugBin)[:slugLengthB32]
		url += slug + "/"
	}

	if p.Zip {
		// Serve contents of zip archive
		rcZip, err := zip.OpenReader(p.Path)
		if err != nil {
			log.Fatalf("Unable to open zip archive: %v", err)
		}
		fs = zipfs.New(rcZip, "onionize")
	} else {
		fileInfo, err := os.Stat(p.Path)
		if err != nil {
			log.Fatalf("Unable to open path: %v", err)
		}
		if fileInfo.IsDir() {
			// Serve a plain directory
			fs = vfs.OS(p.Path)
		} else {
			// Serve just one file in OnionShare-like manner
			abspath, err := filepath.Abs(p.Path)
			if err != nil {
				log.Fatalf("Unable to get absolute path to file")
			}
			dir, file := filepath.Split(abspath)
			m := make(map[string]string)
			m[file] = file
			fs = pickfs.New(vfs.OS(dir), m)
			// Escape URL to be safe and copypasteble
			escapedFilename := strings.Replace(neturl.QueryEscape(file), "+", "%20", -1)
			url += escapedFilename
		}
	}
	// Serve our virtual filesystem
	fileserver := http.FileServer(httpfs.New(fs))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if debug {
			log.Printf("Request for \"%s\"", req.URL)
		}
		err := CheckAndRewriteSlug(req, slug)
		if err != nil {
			if debug {
				log.Print(err)
			}
			err := ResetHTTPConn(&w)
			if err != nil {
				log.Printf("Unable to reset connection: %v", err)
			}
			return
		}
		if debug {
			log.Printf("Rewriting URL to \"%s\"", req.URL)
		}
		fileserver.ServeHTTP(w, req)
	})

	// Connect to a running tor instance
	c, err := bulb.DialURL(p.ControlPath)
	if err != nil {
		log.Fatalf("Failed to connect to control socket: %v", err)
	}
	defer c.Close()

	// See what's really going on under the hood
	c.Debug(debug)

	// Authenticate with the control port
	if err := c.Authenticate(p.ControlPassword); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	// Derive onion service keymaterial from passphrase or generate a new one
	var onionListener net.Listener

	if p.Passphrase != "" {
		privOnionKey, err := onionutil.GenerateOnionKey(onionutil.KeystreamReader([]byte(p.Passphrase), []byte("onionize-keygen")))
		if err != nil {
			log.Fatalf("Unable to generate onion key: %v", err)
		}
		onionListener, err = c.AwaitListener(80, privOnionKey)
	} else {
		onionListener, err = c.AwaitListener(80, nil)
	}
	if err != nil {
		log.Fatalf("Error occured while creating an onion service: %v", err)
	}
	defer onionListener.Close()
	onionHost, _, err := net.SplitHostPort(onionListener.Addr().String())
	if err != nil {
		log.Fatalf("Unable to derive onionID from listener.Addr(): %v", err)
	}
	// Return th link to the service
	linkCh <- ResultLink{URL: fmt.Sprintf("http://%s/%s", onionHost, url), Error: nil}
	// Run webservice
	err = http.Serve(onionListener, nil)
	if err != nil {
		log.Fatalf("Cannot serve HTTP")
	}
}
