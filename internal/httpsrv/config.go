package httpsrv

import (
	"net/http"
	"time"
)

type Config struct {
	// fields from http.Server see https://golang.org/pkg/net/http/#Server
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
}

// NewConfig creates a new config with default values
func NewConfig() Config {
	return Config{
		Addr:              ":8080",
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		ShutdownTimeout:   5 * time.Second,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}
}

func (c *Config) Validate() error {
	return nil
}
