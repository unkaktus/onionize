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
	"flag"
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
	"github.com/nogoegst/terminal"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

type Parameters struct {
	Path	string
	Zip	bool
	Slug	bool
	ControlPath	string
	ControlPassword	string
	Passphrase	string
}

var paramsCh = make(chan Parameters)
var urlCh = make(chan string)

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

func CheckAndRewriteSlug(req *http.Request, slugPrefix string) error {
	if slugPrefix == "" {
		return nil
	}
	reqURL := req.URL.String()
	if len(reqURL) < len(slugPrefix) {
		return fmt.Errorf("URL is too short to have a slug in it")
	}
	if 1 != subtle.ConstantTimeCompare([]byte(slugPrefix), []byte(reqURL[:len(slugPrefix)])) {
		return fmt.Errorf("Wrong slug")
	}
	reqURL = strings.TrimPrefix(reqURL, slugPrefix)
	req.URL, _ = neturl.Parse(reqURL)
	return nil
}

func main() {
	var debugFlag = flag.Bool("debug", false,
		"Show what's happening")
	var zipFlag = flag.Bool("zip", false,
		"Serve zip file contents")
	var passphraseFlag = flag.Bool("p", false,
		"Ask for passphrase to generate onion key")
	var control = flag.String("control-addr", "default://",
		"Set Tor control address to be used")
	var controlPasswd = flag.String("control-passwd", "",
		"Set Tor control auth password")
	flag.Parse()

	if len(flag.Args()) == 0 {
		go guiMain()
	} else {
		go func() {
			p := Parameters{}
			if len(flag.Args()) != 1 {
				log.Fatalf("You should specify exactly one path")
			}
			p.Path = flag.Args()[0]
			p.Zip = *zipFlag
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
			fmt.Println(<-urlCh)
		}()
	}
	debug := *debugFlag
	p := <-paramsCh

	var fs vfs.FileSystem
	var url string
	var slug string
	var slugPrefix string
	if p.Slug {
		slugBin := make([]byte, 10)
		_, err := rand.Read(slugBin)
		if err != nil {
			log.Fatalf("Unable to generate slug: %v", err)
		}
		slug = onionutil.Base32Encode(slugBin)[:16]
		url += slug + "/"
		slugPrefix = "/"+slug
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
		err := CheckAndRewriteSlug(req, slugPrefix)
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
	c, err := bulb.DialURL(*control)
	if err != nil {
		log.Fatalf("Failed to connect to control socket: %v", err)
	}
	defer c.Close()

	// See what's really going on under the hood
	c.Debug(debug)

	// Authenticate with the control port
	if err := c.Authenticate(*controlPasswd); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	// Derive onion service keymaterial from passphrase or generate a new one
	var onionListener net.Listener

	if p.Passphrase != "" {
		privOnionKey, err := onionutil.GenerateOnionKey(onionutil.KeystreamReader([]byte(p.Passphrase), []byte("onionize-keygen")))
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
	onionHost, _, err := net.SplitHostPort(onionListener.Addr().String())
	if err != nil {
		log.Fatalf("Unable to derive onionID from listener.Addr(): %v", err)
	}
	onionID := strings.TrimSuffix(onionHost, ".onion")
	// Wait for service descriptor upload
	c.StartAsyncReader()
	if _, err := c.Request("SETEVENTS HS_DESC"); err != nil {
		log.Fatalf("SETEVENTS HS_DESC has failed: %v", err)
	}
	eventPrefix := fmt.Sprintf("HS_DESC UPLOADED %s", onionID)

	for {
		ev, err := c.NextEvent()
		if err != nil {
			log.Fatalf("NextEvent has failed: %v", err)
		}
		if strings.HasPrefix(ev.Reply, eventPrefix) {
			break
		}
	}
	// Display the link to the service
	urlCh <- fmt.Sprintf("http://%s/%s", onionHost, url)
	// Run webservice
	err = http.Serve(onionListener, nil)
	if err != nil {
		log.Fatalf("Cannot serve HTTP")
	}
}
