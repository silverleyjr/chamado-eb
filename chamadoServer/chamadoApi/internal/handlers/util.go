package handlers

import (
	"net/http"
	"net/url"
	"strings"
)

// decodedHeader reads a header value that the frontend percent-encoded with
// encodeURIComponent, so free text (accents, newlines) survives the trip
// through HTTP headers instead of being mangled or rejected outright.
func decodedHeader(r *http.Request, key string) string {
	raw := r.Header.Get(key)
	if raw == "" {
		return ""
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return strings.TrimSpace(raw)
	}
	return strings.TrimSpace(decoded)
}
