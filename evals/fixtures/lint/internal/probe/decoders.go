package probe

import (
	"database/sql/driver"
	"encoding/json/jsontext"
	"fmt"
)

// Money is a value type stored in a database column: Value stays on the value
// so a Money held in a struct field or a map is still a driver.Valuer, and
// only the decoder takes a pointer.
type Money int64

func (m Money) Value() (driver.Value, error) { return int64(m), nil }

func (m *Money) Scan(src any) error {
	n, ok := src.(int64)
	if !ok {
		return fmt.Errorf("scan Money: unsupported type %T", src)
	}
	*m = Money(n)
	return nil
}

// Flag is the same pair for encoding/json/v2.
type Flag bool

func (f Flag) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.Bool(bool(f)))
}

func (f *Flag) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	*f = Flag(tok.Bool())
	return nil
}
