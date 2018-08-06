package doh

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// UserAgent is the HTTP User-Agent header given to the remote server.
const UserAgent = `doh/1.0.0 (+https://github.com/jamescun/doh)`

// Client is a DNS-over-HTTPS client.
type Client struct {
	// Addr is the address of the DNS-over-HTTPS server.
	Addr *url.URL

	// HTTPClient is the network client through which all DNS-over-HTTPS
	// request will be sent. Defaults to http.DefaultClient if not set
	// (not recommended).
	HTTPClient *http.Client

	// AllowHTTP allows questions to be sent without HTTPS.
	AllowHTTP bool
}

// DefaultClient uses Google DNS and http.DefaultClient
var DefaultClient = &Client{
	Addr: &url.URL{
		Scheme: "https",
		Host:   "dns.google.com",
		Path:   "/resolve",
	},
	HTTPClient: http.DefaultClient,
}

// Do executes a DNS-over-HTTPS query against the configured server. Returned
// is the response from the server, the round-trip time to execute the query
// and any error encountered.
func (c *Client) Do(q *Question) (res *Answer, rtt time.Duration, err error) {
	t1 := time.Now()
	defer func() {
		rtt = time.Now().Sub(t1)
	}()

	if c.Addr == nil {
		err = &ClientError{msg: "no server configured"}
		return
	} else if c.Addr.Scheme != "https" && !c.AllowHTTP {
		err = &ClientError{msg: "https required"}
		return
	}

	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}

	r := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme:   c.Addr.Scheme,
			User:     c.Addr.User,
			Host:     c.Addr.Host,
			Path:     c.Addr.Path,
			RawQuery: q.Values().Encode(),
		},
		Header: http.Header{
			"Accept":     {"application/dns-json"},
			"User-Agent": {UserAgent},
		},
		Host: c.Addr.Host,
	}

	w, err := c.HTTPClient.Do(r)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				err = &ClientError{msg: "server timeout", err: err}
			}
		}

		return
	}
	defer w.Body.Close()

	if w.StatusCode == http.StatusOK {
		var x Answer

		err = unmarshalJSON(w.Body, &x)
		if err != nil {
			return
		}

		res = &x
	} else {
		err = HTTPError(w.StatusCode)
	}

	return
}

// Do executes a DNS-over-HTTPS query against the configured server. Returned
// is the response from the server, the round-trip time to execute the query
// and any error encountered.
func Do(q *Question) (*Answer, time.Duration, error) {
	return DefaultClient.Do(q)
}

// ClientError is returned when there is an error creating a Question or
// connecting to an upstream server.
type ClientError struct {
	msg string
	err error
}

// Cause returns the root cause error, or nil if not configured.
func (ce *ClientError) Cause() error {
	return ce.err
}

func (ce *ClientError) Error() string {
	if ce.err != nil {
		return ce.msg + ": " + ce.err.Error()
	}

	return ce.msg
}

// HTTPError is returned when there is an error unrelated to DNS-over-HTTPS.
type HTTPError int

func (he HTTPError) Error() string {
	return fmt.Sprintf("HTTP Error %d", he)
}
