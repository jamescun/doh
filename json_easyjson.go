// +build easyjson

package doh

import (
	"io"

	"github.com/mailru/easyjson"
)

func marshalJSON(w io.Writer, src easyjson.Marshaler) error {
	_, err := easyjson.MarshalToWriter(src, w)
	return err
}

func unmarshalJSON(r io.Reader, dst easyjson.Unmarshaler) error {
	return easyjson.UnmarshalFromReader(r, dst)
}
