package keyloader

import "github.com/rs/zerolog"

// no logging by default
var log zerolog.Logger

func SetLogger(logger zerolog.Logger) {
	log = logger
}
