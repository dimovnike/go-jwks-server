package httphandler

import "time"

type Config struct {
	CacheMaxAge time.Duration
}

func NewConfig() Config {
	return Config{
		CacheMaxAge: time.Hour,
	}
}
