// Package methods mixes methods the docs check skips with ones it reports.
package methods

import "encoding/json"

// Code is an error code.
type Code int

func (c Code) Error() string               { return "code" }
func (c Code) String() string              { return "code" }
func (c Code) Unwrap() error               { return nil }
func (c Code) Read(p []byte) (int, error)  { return 0, nil }
func (c Code) Write(p []byte) (int, error) { return len(p), nil }

type cache struct{}

func (c *cache) Get(string) string { return "" }

func New() Code { return Code(len(cache{}.name())) }

func (cache) name() string { return "" }

func (c Code) MarshalJSON() ([]byte, error) { return json.Marshal(int(c)) }
