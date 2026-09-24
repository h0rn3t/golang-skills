// Package methods mixes methods the docs check skips with one it reports.
package methods

import "encoding/json"

// Code is an error code.
type Code int

func (c Code) Error() string  { return "code" }
func (c Code) String() string { return "code" }
func (c Code) Unwrap() error  { return nil }

func (c Code) MarshalJSON() ([]byte, error) { return json.Marshal(int(c)) }

type cache struct{}

func (c *cache) Get(string) string { return "" }

func New() Code { return Code(len(cache{}.name())) }

func (cache) name() string { return "" }
