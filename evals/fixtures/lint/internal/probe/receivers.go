package probe

import (
	"encoding/json"
	"strconv"
)

// Counter mixes receivers: go-interfaces says one kind per type.
type Counter struct{ n int }

func (c *Counter) Inc()      { c.n++ }
func (c Counter) Value() int { return c.n }

// Celsius is the go-interfaces exception: encoders and String stay on the
// value, only the decoder takes a pointer.
type Celsius float64

func (c Celsius) String() string { return strconv.FormatFloat(float64(c), 'f', 1, 64) + "C" }

func (c Celsius) MarshalJSON() ([]byte, error) { return json.Marshal(c.String()) }

func (c *Celsius) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	*c = Celsius(f)
	return nil
}
