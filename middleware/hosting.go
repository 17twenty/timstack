package middleware

import (
	"net/http"
	"regexp"
	"strings"
)

var (
	xForwardedScheme = http.CanonicalHeaderKey("X-Forwarded-Scheme")
	xForwardedProto  = http.CanonicalHeaderKey("X-Forwarded-Proto")
	// RFC7239 defines a new "Forwarded: " header designed to replace the
	// existing use of X-Forwarded-* headers.
	// e.g. Forwarded: for=192.0.2.60;proto=https;by=203.0.113.43.
	forwarded = http.CanonicalHeaderKey("Forwarded")
	// Allows for a sub-match for the first instance of scheme (http|https)
	// prefixed by 'proto='. The match is case-insensitive.
	protoRegex = regexp.MustCompile(`(?i)(?:proto=)(https|http)`)
)

func getScheme(r *http.Request) string {
	// Get the scheme
	scheme := r.URL.Scheme

	// Retrieve the scheme from X-Forwarded-Proto.
	if proto := r.Header.Get(xForwardedProto); proto != "" {
		scheme = strings.ToLower(proto)
	} else if proto = r.Header.Get(xForwardedScheme); proto != "" {
		scheme = strings.ToLower(proto)
	} else if proto = r.Header.Get(forwarded); proto != "" {
		// match should contain at least two elements if the protocol was
		// specified in the Forwarded header. The first element will always be
		// the 'proto=' capture, which we ignore. In the case of multiple proto
		// parameters (invalid) we only extract the first.
		if match := protoRegex.FindStringSubmatch(proto); len(match) > 1 {
			scheme = strings.ToLower(match[1])
		}
	}

	return scheme
}

// httpsForwardMiddleware checks for X-Forwarded-Proto and redirects
// http to https
func httpsForwardMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme := getScheme(r)
		// Check for http
		if scheme != "" {
			r.URL.Scheme = scheme
		}

		if scheme != "https" {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
