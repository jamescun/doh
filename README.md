# DNS-over-HTTPS

This package implements a DNS-over-HTTPS client and server in Go. Currently only JSON responses are implemented.

The client is tested against both [Google DNS](https://developers.google.com/speed/public-dns/docs/dns-over-https) and [CloudFlare 1.1.1.1](https://developers.cloudflare.com/1.1.1.1/dns-over-https/) DNS-over-HTTPS implementation.

```sh
go get -u github.com/jamescun/doh
```

## Command Line Utilities

  - [doh](cmd/doh/): human-readable DNS-over-HTTPS command line client


## Libraries

## Client

This package directly exposes a client configured to connect to Google's Public DNS:

```go
answer, rtt, err := doh.Do(&doh.Question{
	Name: "example.org.",
	Type: doh.A,
})
```

If you with to use another DNS-over-HTTPS provider, such as CloudFlare 1.1.1.1, one can be configured:

```go
client := &doh.Client{
	Addr: &url.URL{
		Scheme: "https",
		Host: "cloudflare-dns.com",
		Path: "/dns-query",
	},
}
```


## Server

This package includes a `net/http` compatible server which can be mounted directly under a `http.Server` or with your favourite router.

A simple handler which replies localhost to every question might look like:

```go
func myHandler(q *Question) *Answer {
	if q.Type == doh.A {
		return &Answer{
			Status:   doh.NoError,
			Question: doh.Questions{q},
			Answer:   doh.Records{
				&doh.Record{Name: "example.org", Type: doh.A, TTL: 300, Data: "127.0.0.1"},
			},
		}
	} else if q.Type == doh.AAAA {
		return &Answer{
			Status:   doh.NoError,
			Question: doh.Questions{q},
			Answer:   doh.Records{
				&doh.Record{Name: "example.org", Type: doh.AAAA, TTL: 300, Data: "::1/128"},
			},
		}
	} else {
		return &Answer{
			Status:   doh.NoError,
			Question: doh.Questions{q},
		}
	}
}
```

Handlers are attached to a `Server` object:

```go
dns := &doh.Server{
	Handler: myHandler,
}

h := &http.Server{
	Addr:    "127.0.0.1:443",
	Handler: dns,
}

h.ListenAndServerTLS("cert.pem", "key.pem")
```
