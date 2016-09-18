package auth

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func AuthHandler(validToken string, h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader, ok := r.Header["X-Auth-Token"]
		if !ok || len(authHeader) < 1 {
			noAccessError(w, r)
			return
		}

		if authHeader[0] != validToken {
			noAccessError(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func noAccessError(w http.ResponseWriter, r *http.Request) {
	log.Info("unauthenticated request made")
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}
