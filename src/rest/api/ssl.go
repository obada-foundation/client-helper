package api

import (
	"crypto/tls"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// sslMode defines ssl mode for rest server
type sslMode int8

const (
	// None defines to run http server only
	None sslMode = iota

	// Static defines to run both https and http server. Redirect http to https
	Static
)

// SSLConfig holds all ssl params for rest server
type SSLConfig struct {
	Type sslMode
	Cert string
	Key  string
}

// httpToHTTPSRouter creates new router which does redirect from http to https server  with default middlewares. Used in 'static' ssl mode.
func (s *Rest) httpToHTTPSRouter() chi.Router {
	s.Logger.Debug("create https-to-http redirect routes")
	router := chi.NewRouter()

	router.Handle("/*", s.redirectHandler())
	return router
}

// nolint
func (s *Rest) redirectHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newURL := s.ClientHelperURL + r.URL.Path
		if r.URL.RawQuery != "" {
			newURL += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, newURL, http.StatusTemporaryRedirect)
	})
}

// makeHTTPSServer makes https server for static mode
func (s *Rest) makeHTTPSServer(address string, port int, router http.Handler) *http.Server {
	server := s.makeHTTPServer(address, port, router)
	server.TLSConfig = s.makeTLSConfig()
	return server
}

func (s *Rest) makeTLSConfig() *tls.Config {
	return &tls.Config{
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			// tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			// tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		},
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
			tls.CurveP384,
		},
	}
}
