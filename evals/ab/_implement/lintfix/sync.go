// Package lintfix syncs the partner catalog on a schedule.
package lintfix

import (
	"encoding/json"
	"math/rand/v2"
	"os"
	"time"
)

// Config is the sync job's settings file.
type Config struct {
	Endpoint string        `json:"endpoint"`
	Retries  int           `json:"retries"`
	Base     time.Duration `json:"base"`
}

// LoadConfig reads the settings file at path. A file that is missing or is
// not a valid settings document is an error.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return cfg, nil
}

// Backoff returns how long to wait before retry attempt n (from 0): base
// doubled n times, plus a random jitter below that, so that clients that
// failed together do not retry together.
func Backoff(base time.Duration, n int) time.Duration {
	d := base << n
	return d + time.Duration(rand.Int64N(int64(d)))
}
