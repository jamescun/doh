package doh

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

// Return Codes as defined by IANA.
// https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml
const (
	Success        ReturnCode = 0  // NoError   - No Error
	FormatError               = 1  // FormErr   - Format Error
	ServerFailure             = 2  // ServFail  - Server Failure
	NameError                 = 3  // NXDomain  - Non-Existent Domain
	NotImplemented            = 4  // NotImp    - Not Implemented
	Refused                   = 5  // Refused   - Query Refused
	YXDomain                  = 6  // YXDomain  - Name Exists when it should not
	YXRrset                   = 7  // YXRRSet   - RR Set Exists when it should not
	NXRrset                   = 8  // NXRRSet   - RR Set that should exist does not
	NotAuth                   = 9  // NotAuth   - Server Not Authoritative for zone
	NotZone                   = 10 // NotZone   - Name not contained in zone
	BadSig                    = 16 // BADSIG    - TSIG Signature Failure
	BadVers                   = 16 // BADVERS   - Bad OPT Version
	BadKey                    = 17 // BADKEY    - Key not recognized
	BadTime                   = 18 // BADTIME   - Signature out of time window
	BadMode                   = 19 // BADMODE   - Bad TKEY Mode
	BadName                   = 20 // BADNAME   - Duplicate key name
	BadAlg                    = 21 // BADALG    - Algorithm not supported
	BadTrunc                  = 22 // BADTRUNC  - Bad Truncation
	BadCookie                 = 23 // BADCOOKIE - Bad/missing Server Cookie
)

// ReturnCode is the numerical status in response to a DNS question.
type ReturnCode int

func (rc ReturnCode) String() string {
	switch rc {
	case Success:
		return "NOERROR"
	case FormatError:
		return "FORMERR"
	case ServerFailure:
		return "SERVFAIL"
	case NameError:
		return "NXDOMAIN"
	case NotImplemented:
		return "NOTIMP"
	case Refused:
		return "REFUSED"
	case YXDomain:
		return "YXDOMAIN"
	case YXRrset:
		return "YXRRSET"
	case NXRrset:
		return "NXRRSET"
	case NotAuth:
		return "NOTAUTH"
	case NotZone:
		return "NOTZONE"
	case BadSig:
		return "BADSIG"
	case BadKey:
		return "BADKEY"
	case BadTime:
		return "BADTIME"
	case BadMode:
		return "BADMODE"
	case BadAlg:
		return "BADALG"
	case BadTrunc:
		return "BADTRUNC"
	case BadCookie:
		return "BADCOOKIE"
	default:
		return "UNKNOWN"
	}
}

var rrNameToInt = map[string]RecordType{
	"A":          A,
	"NS":         NS,
	"CNAME":      CNAME,
	"SOA":        SOA,
	"PTR":        PTR,
	"HINFO":      HINFO,
	"MX":         MX,
	"TXT":        TXT,
	"RP":         RP,
	"AFSDB":      AFSDB,
	"SIG":        SIG,
	"KEY":        KEY,
	"AAAA":       AAAA,
	"LOC":        LOC,
	"SRV":        SRV,
	"NAPTR":      NAPTR,
	"KX":         KX,
	"CERT":       CERT,
	"DNAME":      DNAME,
	"OPT":        OPT,
	"APL":        APL,
	"DS":         DS,
	"SSHFP":      SSHFP,
	"IPSECKEY":   IPSECKEY,
	"RRSIG":      RRSIG,
	"NSEC":       NSEC,
	"DNSKEY":     DNSKEY,
	"DHCID":      DHCID,
	"NSEC3":      NSEC3,
	"NSEC3PARAM": NSEC3PARAM,
	"TLSA":       TLSA,
	"HIP":        HIP,
	"CDS":        CDS,
	"CDNSKEY":    CDNSKEY,
	"OPENPGPKEY": OPENPGPKEY,
	"SPF":        SPF,
	"TKEY":       TKEY,
	"TSIG":       TSIG,
	"IXFR":       IXFR,
	"AXFR":       AXFR,
	"ANY":        ANY,
	"URI":        URI,
	"CAA":        CAA,
	"TA":         TA,
	"DLV":        DLV,
}

// RecordTypeFromString returns the numerical Record Type from its
// stringified representation.
func RecordTypeFromString(name string) RecordType {
	rr, ok := rrNameToInt[name]
	if !ok {
		return -1
	}

	return rr
}

// Record Types as defined by IANA.
const (
	A          RecordType = 1
	NS                    = 2
	CNAME                 = 5
	SOA                   = 6
	PTR                   = 12
	HINFO                 = 13
	MX                    = 15
	TXT                   = 16
	RP                    = 17
	AFSDB                 = 18
	SIG                   = 24
	KEY                   = 25
	AAAA                  = 28
	LOC                   = 29
	SRV                   = 33
	NAPTR                 = 35
	KX                    = 36
	CERT                  = 37
	DNAME                 = 39
	OPT                   = 41
	APL                   = 42
	DS                    = 43
	SSHFP                 = 44
	IPSECKEY              = 45
	RRSIG                 = 46
	NSEC                  = 47
	DNSKEY                = 48
	DHCID                 = 49
	NSEC3                 = 50
	NSEC3PARAM            = 51
	TLSA                  = 52
	HIP                   = 55
	CDS                   = 59
	CDNSKEY               = 60
	OPENPGPKEY            = 61
	SPF                   = 99
	TKEY                  = 249
	TSIG                  = 250
	IXFR                  = 251
	AXFR                  = 252
	ANY                   = 255
	URI                   = 256
	CAA                   = 257
	TA                    = 32768
	DLV                   = 32769
)

// RecordType is the numerical representation of a DNS records type.
type RecordType int

// UnmarshalJSON supports unmarshaling both numerical and stringified
// DNS record types into their numerical form.
func (rr *RecordType) UnmarshalJSON(b []byte) error {
	if len(b) < 1 {
		return errors.New("unexpected end of JSON input")
	}

	if b[0] == '"' && b[len(b)-1] == '"' {
		// TODO(jc): support lowercase rrtype
		if len(b) < 3 {
			*rr = 0
			return nil
		}

		x, ok := rrNameToInt[string(b[1:len(b)-1])]
		if !ok {
			return errors.New("unknown record type")
		}

		*rr = x
	} else {
		var x int

		err := json.Unmarshal(b, &x)
		if err != nil {
			return err
		}

		*rr = RecordType(x)
	}

	return nil
}

// String returns the string representation of a numerical Record Type.
func (rr RecordType) String() string {
	for k, v := range rrNameToInt {
		if rr == v {
			return k
		}
	}

	return "unknown(" + strconv.Itoa(int(rr)) + ")"
}

// Record is a single DNS record that is returned as part of an Answer.
//
//easyjson:json
type Record struct {
	Name string     `json:"name"`
	Type RecordType `json:"type"`
	TTL  int        `json:"TTL"`
	Data string     `json:"data"`
}

// Records contains one-or-more DNS records.
type Records []*Record

// Question contains the incoming request to a server, or the outgoing request
// to a client. Only the Name field is required.
//
//easyjson:json
type Question struct {
	// Name is the hostname to request from the upstream server. It MAY be
	// fully qualified.
	Name string `json:"name"`

	// Type is the numerical record type to specifically request.
	Type RecordType `json:"type"`

	// DisableDNSSEC is a request by the client to skip DNSSEC validation.
	DisableDNSSEC bool `json:"CD"`

	// EDNSClientSubnet is the optional IPv4/IPv6 subnet of the client when
	// resolving on behalf of another client.
	EDNSClientSubnet string `json:"edns_client_subnet,omitempty"`
}

// Questions contains one-or-more questions.
type Questions []*Question

// QuestionFromValues unmarshals a Question object from a set of URL parameters.
func QuestionFromValues(q url.Values) Question {
	rr, _ := strconv.Atoi(q.Get("type"))
	if rr < 1 {
		rr = 1
	}

	req := Question{
		Name:             q.Get("name"),
		Type:             RecordType(rr),
		EDNSClientSubnet: q.Get("edns_client_subnet"),
	}

	req.DisableDNSSEC, _ = strconv.ParseBool(q.Get("cd"))

	return req
}

// Values returns the Question as a set of URL parameters.
func (r Question) Values() url.Values {
	q := url.Values{
		"name": {r.Name},
	}

	if r.Type > 0 {
		q["type"] = []string{strconv.Itoa(int(r.Type))}
	} else {
		q["type"] = []string{"1"}
	}

	if r.DisableDNSSEC {
		q["cd"] = []string{"1"}
	}

	if r.EDNSClientSubnet != "" {
		q["edns_client_subnet"] = []string{r.EDNSClientSubnet}
	}

	return q
}

// Answer is the full response to be returned to the client in
// response to a Request.
//
//easyjson:json
type Answer struct {
	// Status is the DNS Response Code (RCODE) as defined by IANA
	// https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-6
	Status ReturnCode `json:"Status"`

	// Truncated indicates if the truncate bit was set on the response from the
	// upstream DNS server, meaning the reply was bigger than a single
	// UDP or TCP packet.
	Truncated bool `json:"TC"`

	RecursionDesired   bool `json:"RD"`
	RecursionAvailable bool `json:"RA"`

	// DNSSECValidated indicates all response data was validated with DNSSEC.
	DNSSECValidated bool `json:"AD"`

	// DNSSECDisabled indicates the client requested DNSSEC validation to be
	// disabled.
	DNSSECDisabled bool `json:"CD"`

	Question   Questions `json:"Question"`
	Answer     Records   `json:"Answer"`
	Authority  Records   `json:"Authority,omitempty"`
	Additional Records   `json:"Additional,omitempty"`
	Comment    string    `json:"Comment,omitempty"`

	EdnsClientSubnet string `json:"edns_client_subnet,omitempty"`
}

// IsFQDN returns true if the given name is fully qualified.
func IsFQDN(name string) bool {
	if len(name) > 1 && name[len(name)-1] == '.' {
		return true
	}

	return false
}

// FQDN returns the given name as a FQDN if not already.
func FQDN(name string) string {
	if !IsFQDN(name) {
		return name + "."
	}

	return name
}
