package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

func New(config Config) (zerolog.Logger, error) {
	if err := config.Validate(); err != nil {
		return zerolog.Logger{}, err
	}

	var file *os.File
	switch config.File {
	case "stdout":
		file = os.Stdout
	case "stderr":
		file = os.Stderr
	default:
		return zerolog.Logger{}, fmt.Errorf("invalid log file: %s", config.File)
	}

	logger := zerolog.New(file)

	switch config.Level {
	case "debug":
		logger = logger.Level(zerolog.DebugLevel)
	case "info":
		logger = logger.Level(zerolog.InfoLevel)
	case "warn":
		logger = logger.Level(zerolog.WarnLevel)
	case "error":
		logger = logger.Level(zerolog.ErrorLevel)
	case "fatal":
		logger = logger.Level(zerolog.FatalLevel)
	case "panic":
		logger = logger.Level(zerolog.PanicLevel)
	case "no":
		logger = logger.Level(zerolog.NoLevel)
	case "disabled":
		logger = logger.Level(zerolog.Disabled)
	case "trace":
		logger = logger.Level(zerolog.TraceLevel)
	}

	if config.Timestamp {
		logger = logger.With().Timestamp().Logger()
	}

	if config.Caller {
		logger = logger.With().Caller().Logger()
	}

	if config.Stack {
		logger = logger.With().Stack().Logger()
	}

	if config.Console {
		logger = logger.Output(zerolog.ConsoleWriter{Out: file, NoColor: config.NoColor})
	}

	return logger, nil
}
