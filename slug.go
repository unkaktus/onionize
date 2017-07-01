package onionize

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
)

func checkSlug(req *http.Request, slug string) error {
	if slug == "" {
		return nil
	}
	shost := strings.Split(req.Host, ".")
	if len(shost) < 3 {
		return fmt.Errorf("hostname is too short")
	}
	if 1 != subtle.ConstantTimeCompare([]byte(slug), []byte(shost[len(shost)-3])) {
		return fmt.Errorf("wrong slug")
	}
	return nil
}

func SubdomainSluggedHandler(h http.Handler, slug string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := checkSlug(req, slug)
		if err != nil {
			http.NotFound(w, req)
			return
		}
		h.ServeHTTP(w, req)
	})
	return mux
}
