package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Route mapira prefiks putanje na ciljni mikroservis.
type Route struct {
	Prefix string
	Target string
}

// New pravi reverse proxy koji skida "/api" prefiks i rutira zahtev
// ka odgovarajućem servisu na osnovu prefiksa putanje.
func New(routes []Route) (http.Handler, error) {
	type entry struct {
		prefix string
		proxy  *httputil.ReverseProxy
	}
	entries := make([]entry, 0, len(routes))
	for _, rt := range routes {
		target, err := url.Parse(rt.Target)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry{prefix: rt.Prefix, proxy: httputil.NewSingleHostReverseProxy(target)})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, e := range entries {
			if strings.HasPrefix(r.URL.Path, e.prefix) {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
				e.proxy.ServeHTTP(w, r)
				return
			}
		}
		http.NotFound(w, r)
	}), nil
}
