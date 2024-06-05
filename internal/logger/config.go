package logger

import "fmt"

type Config struct {
	Level     string // logging level
	Timestamp bool   // show timestamp
	Caller    bool   // show caller file and line number
	Stack     bool   // show stack info
	Console   bool   // human friendly logs on console
	NoColor   bool   // do not show color on console
	File      string // log file, stdout or stderr
}

// NewConfig creates a new config with default values
func NewConfig() Config {
	return Config{
		Level:     "error",
		Timestamp: true,
		Caller:    true,
		File:      "stderr",
	}
}

func (c *Config) Validate() error {
	switch c.Level {
	case "debug", "info", "warn", "error", "fatal", "panic", "no", "disabled", "trace":
	default:
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	return nil
}
