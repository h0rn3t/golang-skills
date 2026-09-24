package probe

import "encoding/json"

// Decode returns the object in b.
func Decode(b []byte) map[string]any {
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}
