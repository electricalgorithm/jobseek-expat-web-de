package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
)

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("url")

	// Optional: Support base64 encoded URLs to make them look cleaner and avoid email scanners following them immediately
	// If query param "b64" is present, decode it.
	if r.URL.Query().Get("data") != "" {
		decoded, err := base64.URLEncoding.DecodeString(r.URL.Query().Get("data"))
		if err == nil {
			target = string(decoded)
		}
	}

	if target == "" {
		http.Error(w, "Missing redirection target", http.StatusBadRequest)
		return
	}

	// Validate URL scheme to prevent open redirect vulnerabilities to non-http(s)
	u, err := url.Parse(target)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		http.Error(w, "Invalid redirection target", http.StatusBadRequest)
		return
	}

	log.Printf("[Redirect] Redirecting user to: %s", target)
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}
