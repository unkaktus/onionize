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

// TODO: make it less specific
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
