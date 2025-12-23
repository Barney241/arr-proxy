// Package usecases contains the application's use cases.
package usecases

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// ProxyUseCase is the use case for proxying requests.
type ProxyUseCase struct {
	transport *http.Transport
}

// NewProxyUseCase creates a new ProxyUseCase with configured timeouts.
func NewProxyUseCase() *ProxyUseCase {
	return &ProxyUseCase{
		transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
		},
	}
}

// ServeHTTP proxies the request to the target URL.
func (uc *ProxyUseCase) ServeHTTP(w http.ResponseWriter, r *http.Request, targetURL *url.URL, apiKey string) {
	proxy := &httputil.ReverseProxy{
		Transport: uc.transport,
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path, req.URL.RawPath = joinURLPath(targetURL, req.URL)
			if targetURL.RawQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetURL.RawQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetURL.RawQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
			req.Header.Set("X-Api-Key", apiKey)
			req.Host = targetURL.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("Proxy error", "error", err, "path", r.URL.Path, "target", targetURL.Host)
			http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(w, r)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawPath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but works on RawPath
	apath := a.Path
	if a.RawPath != "" {
		apath = a.RawPath
	}
	bpath := b.Path
	if b.RawPath != "" {
		bpath = b.RawPath
	}

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	if aslash && bslash {
		return singleJoiningSlash(a.Path, b.Path), apath + bpath[1:]
	} else if !aslash && !bslash {
		return singleJoiningSlash(a.Path, b.Path), apath + "/" + bpath
	}
	return singleJoiningSlash(a.Path, b.Path), apath + bpath
}