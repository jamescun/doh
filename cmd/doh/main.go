package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jamescun/doh"
)

// command-line flags
var (
	Server    = flag.String("server", "https://dns.google.com/resolve", "url of dns-over-https server")
	Timeout   = flag.Duration("timeout", 30*time.Second, "query timeout")
	AllowHTTP = flag.Bool("allow-http", false, "allow questions over HTTP")
	JSON      = flag.Bool("json", false, "output json")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		configError("usage: doh [options] <record type> <name>")
	}

	addr, err := url.Parse(*Server)
	if err != nil {
		configError("invalid server: %s", err)
	}

	client := &doh.Client{
		Addr: addr,
		HTTPClient: &http.Client{
			Timeout: *Timeout,
		},
		AllowHTTP: *AllowHTTP,
	}

	answer, rtt, err := client.Do(&doh.Question{
		Name: args[1],
		Type: doh.RecordTypeFromString(args[0]),
	})

	if err != nil {
		runtimeError("could not query server: %s", err)
	}

	if *JSON {
		json.NewEncoder(os.Stdout).Encode(answer)
	} else {
		fmt.Printf("Return Code: %s\n", answer.Status)
		fmt.Printf("Truncated: %t\n", answer.Truncated)
		fmt.Printf("Recursion Desired / Available: %t / %t\n", answer.RecursionDesired, answer.RecursionAvailable)
		fmt.Printf("DNSSEC Disabled / Validated: %t / %t\n\n", answer.DNSSECDisabled, answer.DNSSECValidated)

		fmt.Println("Answer:")
		for _, record := range answer.Answer {
			fmt.Printf("  %s   %s  %d  '%s'\n", record.Name, record.Type, record.TTL, record.Data)
		}

		if len(answer.Additional) > 0 {
			fmt.Println("\nAdditional:")
			for _, record := range answer.Answer {
				fmt.Printf("  %s  %s  %d  '%s'\n", record.Name, record.Type, record.TTL, record.Data)
			}
		}

		if answer.Comment != "" {
			fmt.Printf("\nComment: %s\n", answer.Comment)
		}

		fmt.Printf("\nQuery time: %s\n", rtt)
	}

}

func configError(format string, args ...interface{}) {
	fmt.Printf("config error: "+format+"\n", args...)
	os.Exit(2)
}

func runtimeError(format string, args ...interface{}) {
	fmt.Printf("runtime error: "+format+"\n", args...)
	os.Exit(1)
}
