// proxy.go - reverse proxy to another HTTP server.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionize

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func onionReverseHTTPProxy(target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		req.Header.Set("User-Agent", "onionize")
		log.Printf("%v", req.URL)
	}
	return &httputil.ReverseProxy{Director: director}
}
