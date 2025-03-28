package auth

import (
	"net/http"
)

// Credentials represents the authentication credentials
type Credentials struct {
	Username string
	Password string
}

// BasicAuth middleware handles HTTP Basic Authentication
type BasicAuth struct {
	credentials Credentials
}

// NewBasicAuth creates a new BasicAuth middleware with the given credentials
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		credentials: Credentials{
			Username: username,
			Password: password,
		},
	}
}

// Middleware returns a http.HandlerFunc that performs basic auth
func (ba *BasicAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			ba.unauthorized(w)
			return
		}

		if !ba.validateCredentials(username, password) {
			ba.unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateCredentials checks if the provided credentials match
func (ba *BasicAuth) validateCredentials(username, password string) bool {
	return username == ba.credentials.Username && password == ba.credentials.Password
}

// unauthorized sends an unauthorized response
func (ba *BasicAuth) unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
