package doh

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGoogle(t *testing.T) {
	google := &Client{
		Addr: &url.URL{
			Scheme: "https",
			Host:   "dns.google.com",
			Path:   "/resolve",
		},
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	tests := []struct {
		Name string

		Question *Question
		Answer   *Answer
		Error    error
	}{
		{
			"A:example.org",
			&Question{Name: "example.org.", Type: A},
			&Answer{
				Status:             Success,
				RecursionDesired:   true,
				RecursionAvailable: true,
				DNSSECValidated:    true,
				Question: Questions{
					&Question{Name: "example.org.", Type: A},
				},
				Answer: Records{
					&Record{
						Name: "example.org.",
						Type: A,
						TTL:  6792,
						Data: "93.184.216.34",
					},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			answer, _, err := google.Do(test.Question)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Answer, answer)
				}
			} else {
				assert.Equal(t, test.Error, err)
			}
		})
	}
}

func TestCloudFlare(t *testing.T) {
	cloudflare := &Client{
		Addr: &url.URL{
			Scheme: "https",
			Host:   "cloudflare-dns.com",
			Path:   "/dns-query",
		},
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	tests := []struct {
		Name string

		Question *Question
		Answer   *Answer
		Error    error
	}{
		{
			"A:example.org",
			&Question{Name: "example.org.", Type: A},
			&Answer{
				Status:             Success,
				RecursionDesired:   true,
				RecursionAvailable: true,
				DNSSECValidated:    true,
				Question: Questions{
					&Question{Name: "example.org.", Type: A},
				},
				Answer: Records{
					&Record{
						Name: "example.org.",
						Type: A,
						TTL:  6792,
						Data: "93.184.216.34",
					},
				},
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			answer, _, err := cloudflare.Do(test.Question)

			if test.Error == nil {
				if assert.NoError(t, err) {
					assert.Equal(t, test.Answer, answer)
				}
			} else {
				assert.Equal(t, test.Error, err)
			}
		})
	}
}
