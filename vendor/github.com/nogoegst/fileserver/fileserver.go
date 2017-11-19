// fileserver.go - create an HTTP server from directories, files and zips.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of fileserver, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package fileserver

import (
	"archive/zip"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/nogoegst/pickfs"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

// splitQuoted splits s by sep if it is found outside substring
// quoted by quote.
func splitQuoted(s string, quote, sep rune) (splitted []string) {
	quoteFlag := false
NewSubstring:
	for i, c := range s {
		if c == quote {
			quoteFlag = !quoteFlag
		}
		if c == sep && !quoteFlag {
			splitted = append(splitted, s[:i])
			s = s[i+1:]
			goto NewSubstring
		}
	}
	return append(splitted, s)
}

const delimeter = ';'

func JoinPathspec(paths []string) string {
	return strings.Join(paths, string(delimeter))
}

func parsePathspec(pathspec string) (map[string]string, error) {
	aliasmap := make(map[string]string)
	paths := splitQuoted(pathspec, '"', delimeter)
	for _, path := range paths {
		spath := strings.Split(path, ":")
		var alias string
		switch len(spath) {
		case 1:
			_, alias = filepath.Split(filepath.Clean(spath[0]))
			if alias == "." && len(paths) != 1 {
				return nil, errors.New("current working dir doesnt't have an alias")
			}
		case 2:
			alias = spath[1]
		default:
			return nil, errors.New("invalid filespec: too many delimeters")
		}
		abs, err := filepath.Abs(spath[0])
		if err != nil {
			return nil, err
		}
		alias = filepath.Clean(alias)
		aliasmap[alias] = abs
	}
	return aliasmap, nil
}

// Creates new handler that serves files from path. Serves from
// zip archive if zipOn is set.
func New(pathspec string, zipOn, debug bool) (http.Handler, error) {
	var fs vfs.FileSystem
	var aliasmap map[string]string
	traverseLonelyPath := true
	if zipOn {
		// Serve contents of zip archive
		rcZip, err := zip.OpenReader(pathspec)
		if err != nil {
			return nil, fmt.Errorf("Unable to open zip archive: %v", err)
		}
		fs = zipfs.New(rcZip, "zipfs")
	} else {
		var err error
		aliasmap, err = parsePathspec(pathspec)
		if err != nil {
			return nil, err
		}
		fs = pickfs.New(vfs.OS(""), aliasmap)
		if _, ok := aliasmap["."]; ok {
			fs = vfs.OS(".")
			traverseLonelyPath = false
		}
	}
	fileserver := http.FileServer(httpfs.New(fs))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if debug {
			log.Printf("Request for \"%s\"", req.URL)
		}
		// Traverse lonely path
		if traverseLonelyPath && req.URL.String() == "/" {
			lpath := "/"
			for {
				fi, err := fs.ReadDir(lpath)
				if err != nil || len(fi) != 1 {
					break
				}
				lpath = filepath.Join(lpath, fi[0].Name())
			}
			if lpath != "/" {
				http.Redirect(w, req, lpath, http.StatusFound)
				return
			}
		}
		fileserver.ServeHTTP(w, req)
	})
	return mux, nil
}

// Same as New, but attaches a server to listener l.
func Serve(l net.Listener, path string, zipOn, debug bool) error {
	fs, err := New(path, zipOn, debug)
	if err != nil {
		return err
	}
	s := http.Server{Handler: fs}
	return s.Serve(l)
}
