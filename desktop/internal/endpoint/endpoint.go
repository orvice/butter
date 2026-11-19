// Package endpoint normalizes the server address a user types into the
// client: the scheme defaults to https, a trailing slash is dropped, and
// anything a ConnectRPC base URL cannot carry (user info, query, fragment)
// is rejected instead of silently ignored.
package endpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ErrInvalid reports an address the client cannot use.
var ErrInvalid = errors.New("endpoint: invalid server address")

// Server is a normalized Butter server address.
type Server struct {
	base url.URL
}

// Parse normalizes raw. It accepts a host ("butter.example.com"), an origin
// with a port, or a deployment under a path prefix.
func Parse(raw string) (Server, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Server{}, fmt.Errorf("%w: empty address", ErrInvalid)
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return Server{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return Server{}, fmt.Errorf("%w: scheme %q is not http or https", ErrInvalid, u.Scheme)
	}
	if u.Host == "" || u.Hostname() == "" {
		return Server{}, fmt.Errorf("%w: missing host", ErrInvalid)
	}
	if u.User != nil {
		return Server{}, fmt.Errorf("%w: credentials in the address are not accepted", ErrInvalid)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return Server{}, fmt.Errorf("%w: query and fragment are not part of a server address", ErrInvalid)
	}
	if u.Path == "/" {
		u.Path = ""
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = ""
	return Server{base: *u}, nil
}

// Valid reports whether s came from Parse.
func (s Server) Valid() bool { return s.base.Hostname() != "" }

// String is the normalized address, without a trailing slash.
func (s Server) String() string { return s.base.String() }

// APIBase is the base URL of the ConnectRPC and AG-UI routes.
func (s Server) APIBase() string { return s.String() + "/api" }

// Key identifies the server in stored credentials and cached state, so a
// token kept for one server is never offered to another.
func (s Server) Key() string {
	sum := sha256.Sum256([]byte(s.String()))
	return hex.EncodeToString(sum[:8])
}
