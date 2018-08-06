// +build !easyjson

package doh

import (
	"encoding/json"
	"io"
)

func marshalJSON(w io.Writer, src interface{}) error {
	return json.NewEncoder(w).Encode(src)
}

func unmarshalJSON(r io.Reader, dst interface{}) error {
	return json.NewDecoder(r).Decode(dst)
}
