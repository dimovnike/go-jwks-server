package keyloader

import (
	"errors"
	"time"
)

type Config struct {
	// the directory to load the keys from
	Dir string

	// set to 0 to disable watching
	WatchInterval time.Duration

	// fail on error, actually return the error, otherwise just log it
	FailOnError bool
}

// NewConfig creates a new config with default values
func NewConfig() Config {
	return Config{
		Dir:           "./keys",
		WatchInterval: 1 * time.Second,
		FailOnError:   false,
	}
}

func (c *Config) Validate() error {
	if c.Dir == "" {
		return errors.New("key-dir is required")
	}

	return nil
}

func (c *Config) WatchOn() bool {
	return c.WatchInterval > 0
}
