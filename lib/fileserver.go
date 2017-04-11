// fileserver.go - onionize directories, files and zips.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionize

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nogoegst/pickfs"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

func FileServer(path string, zipOn, debug bool) (http.Handler, error) {
	var fs vfs.FileSystem
	var filename string

	if zipOn {
		// Serve contents of zip archive
		rcZip, err := zip.OpenReader(path)
		if err != nil {
			return nil, fmt.Errorf("Unable to open zip archive: %v", err)
		}
		fs = zipfs.New(rcZip, "onionize")
	} else {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("Unable to open path: %v", err)
		}
		if fileInfo.IsDir() {
			// Serve a plain directory
			fs = vfs.OS(path)
		} else {
			// Serve just one file
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, fmt.Errorf("Unable to get absolute path to file")
			}
			var dir string
			dir, filename = filepath.Split(abspath)
			m := make(map[string]string)
			m[filename] = filename
			fs = pickfs.New(vfs.OS(dir), m)
		}
	}
	fileserver := http.FileServer(httpfs.New(fs))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if debug {
			log.Printf("Request for \"%s\"", req.URL)
		}
		// Redirect roots to the file itself
		if req.URL.String() == "/" && filename != "" {
			http.Redirect(w, req, "/"+filename, http.StatusFound)
			return
		}
		fileserver.ServeHTTP(w, req)
	})
	return mux, nil
}
