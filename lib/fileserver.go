// fileserver.go - onionize directories, files and zips.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionize

import (
	"archive/zip"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/nogoegst/onionutil"
	"github.com/nogoegst/pickfs"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

func checkAndRewriteSlug(req *http.Request, slug string) error {
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
	req.URL, _ = url.Parse(reqURL)
	return nil
}

func FileServer(path string, slugOn, zipOn, debug bool) (handler http.Handler, link string, err error) {
	var fs vfs.FileSystem
	var slug string
	if slugOn {
		slugBin := make([]byte, (slugLengthB32*5)/8+1)
		_, err := rand.Read(slugBin)
		if err != nil {
			return nil, "", fmt.Errorf("Unable to generate slug: %v", err)
		}
		slug = onionutil.Base32Encode(slugBin)[:slugLengthB32]
		link = slug + "/"
	}

	if zipOn {
		// Serve contents of zip archive
		rcZip, err := zip.OpenReader(path)
		if err != nil {
			return nil, "", fmt.Errorf("Unable to open zip archive: %v", err)
		}
		fs = zipfs.New(rcZip, "onionize")
	} else {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, "", fmt.Errorf("Unable to open path: %v", err)
		}
		if fileInfo.IsDir() {
			// Serve a plain directory
			fs = vfs.OS(path)
		} else {
			// Serve just one file in OnionShare-like manner
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, "", fmt.Errorf("Unable to get absolute path to file")
			}
			dir, filename := filepath.Split(abspath)
			m := make(map[string]string)
			m[filename] = filename
			fs = pickfs.New(vfs.OS(dir), m)
			link += filename
		}
	}
	// Serve our virtual filesystem
	fileserver := http.FileServer(httpfs.New(fs))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if debug {
			log.Printf("Request for \"%s\"", req.URL)
		}
		err := checkAndRewriteSlug(req, slug)
		if err != nil {
			if debug {
				log.Print(err)
			}
			http.NotFound(w, req)
			return
		}
		if req.URL.String() == "" { // empty root path
			http.Redirect(w, req, "/"+slug+"/", http.StatusFound)
			return
		}
		if debug {
			log.Printf("Rewriting URL to \"%s\"", req.URL)
		}
		fileserver.ServeHTTP(w, req)
	})
	return mux, link, nil
}
