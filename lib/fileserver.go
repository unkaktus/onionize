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
	"os"
	"path/filepath"
	"strings"

	"github.com/nogoegst/onionutil"
	"github.com/nogoegst/pickfs"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

func checkSlug(req *http.Request, slug string) error {
	if slug == "" {
		return nil
	}
	shost := strings.Split(req.Host, ".")
	if len(shost) != 3 {
		return fmt.Errorf("Wrong hostname to have a slug")
	}
	if len(shost[0]) < slugLength {
		return fmt.Errorf("Subdomain is too short to have a slug in it")
	}
	if 1 != subtle.ConstantTimeCompare([]byte(slug), []byte(shost[0][:len(slug)])) {
		return fmt.Errorf("Wrong slug")
	}
	return nil
}

func generateSlug() (string, error) {
	slugBin := make([]byte, (slugLength*5)/8+1)
	_, err := rand.Read(slugBin)
	if err != nil {
		return "", err
	}
	return onionutil.Base32Encode(slugBin)[:slugLength], nil
}

func FileServer(path string, slugOn, zipOn, debug bool) (handler http.Handler, slug string, err error) {
	var fs vfs.FileSystem
	var filename string

	if slugOn {
		slug, err = generateSlug()
		if err != nil {
			return nil, "", fmt.Errorf("Unable to generate slug: %v", err)
		}
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
			// Serve just one file
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, "", fmt.Errorf("Unable to get absolute path to file")
			}
			var dir string
			dir, filename = filepath.Split(abspath)
			m := make(map[string]string)
			m[filename] = filename
			fs = pickfs.New(vfs.OS(dir), m)
		}
	}
	// Serve our virtual filesystem
	fileserver := http.FileServer(httpfs.New(fs))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if debug {
			log.Printf("Request for \"%s\"", req.URL)
		}
		err := checkSlug(req, slug)
		if err != nil {
			http.NotFound(w, req)
			if debug {
				log.Print(err)
			}
			return
		}
		// Redirect roots to the file itself
		if req.URL.String() == "/" && filename != "" {
			http.Redirect(w, req, "/"+filename, http.StatusFound)
			return
		}
		fileserver.ServeHTTP(w, req)
	})
	return mux, slug, nil
}
