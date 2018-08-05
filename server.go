package doh

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HandlerFunc is given an incoming DNS-over-HTTPS request and is expected to
// return a response, if the response is nil a ServFail rcode will be returned
// to the client.
type HandlerFunc func(*Question) *Answer

// Server configures handling of DNS-over-HTTPS requests and exposes a
// net/http compatible server.
type Server struct {
	// Handler is invoked for every valid DNS-over-HTTPS request.
	Handler HandlerFunc

	// AllowHTTP disables refusing to answer requests that did not come
	// over HTTPS.
	AllowHTTP bool
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if r.Header.Get("Accept") == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if strings.HasPrefix(r.Header.Get("Accept"), "application/dns-json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	} else if s.Handler == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if r.URL.Scheme != "https" && !s.AllowHTTP {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	req := QuestionFromValues(r.URL.Query())

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/dns-json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	res := s.Handler(&req)

	if res == nil {
		res = &Answer{
			Status: ServerFailure,
		}
	}

	json.NewEncoder(w).Encode(res)
}
