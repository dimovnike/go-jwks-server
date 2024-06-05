package httphandler

import "time"

type Config struct {
	KeysEndpoint string
	CacheMaxAge  time.Duration
}

func NewConfig() Config {
	return Config{
		KeysEndpoint: "/keys",
		CacheMaxAge:  time.Hour,
	}
}
