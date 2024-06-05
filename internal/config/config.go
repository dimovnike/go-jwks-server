package config

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"go-jwks-server/internal/httphandler"
	"go-jwks-server/internal/httpsrv"
	"go-jwks-server/internal/keyloader"
	"go-jwks-server/internal/logger"
	"html/template"
	"os"
	"strings"
)

const envVarPrefix = "GO_JWKS_SERVER_"

//go:embed usage.tpl
var usageTemplate string

type Config struct {
	Logger      logger.Config
	Keyloader   keyloader.Config
	Httpsrv     httpsrv.Config
	Httphandler httphandler.Config

	PrintConfig bool
}

func (c Config) String() string {
	j, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(j)
}

func New() (Config, error) {
	config := Config{
		Logger:      logger.NewConfig(),
		Keyloader:   keyloader.NewConfig(),
		Httpsrv:     httpsrv.NewConfig(),
		Httphandler: httphandler.NewConfig(),
	}

	flag.Usage = func() {
		var exampleFlag string
		flag.VisitAll(func(f *flag.Flag) {
			if exampleFlag == "" {
				exampleFlag = f.Name
			}
		})

		if exampleFlag == "" {
			exampleFlag = "example-flag"
		}

		t := template.Must(template.New("usage").Parse(usageTemplate))

		err := t.Execute(os.Stderr, map[string]string{
			"envVarPrefix":  envVarPrefix,
			"exampleFlag":   exampleFlag,
			"exampleEnvVar": envVarName(exampleFlag),
		})
		if err != nil {
			// no logging support here
			panic(err)
		}

		fmt.Fprint(os.Stderr, "\n")

		flag.PrintDefaults()
	}

	// logger config

	flag.StringVar(&config.Logger.Level, "log-level", config.Logger.Level,
		"logging level (debug, info, warn, error, fatal, panic, no, disabled, trace)")

	flag.BoolVar(&config.Logger.Timestamp, "log-timestamp", config.Logger.Timestamp,
		"show timestamp")

	flag.BoolVar(&config.Logger.Caller, "log-caller", config.Logger.Caller,
		"show caller file and line number")

	flag.BoolVar(&config.Logger.Stack, "log-stack", config.Logger.Stack,
		"show stack info")

	flag.BoolVar(&config.Logger.Console, "log-console", config.Logger.Console,
		"human friendly logs on console")

	flag.BoolVar(&config.Logger.NoColor, "log-no-color", config.Logger.NoColor,
		"do not show color on console")

	flag.StringVar(&config.Logger.File, "log-file", config.Logger.File,
		"log file, stdout or stderr")

	// keyloader config

	flag.StringVar(&config.Keyloader.Dir, "key-dir", config.Keyloader.Dir,
		"the directory to load the keys from")

	flag.DurationVar(&config.Keyloader.WatchInterval, "dir-watch-interval", config.Keyloader.WatchInterval,
		"the interval to check the key directory for changes, set to 0 to disable watching")

	flag.BoolVar(&config.Keyloader.FailOnError, "exit-on-error", config.Keyloader.FailOnError,
		"exit if loading keys fails")

	// http config

	flag.StringVar(&config.Httpsrv.Addr, "http-addr", config.Httpsrv.Addr,
		"the address to listen on")

	flag.DurationVar(&config.Httpsrv.ReadTimeout, "http-read-timeout", config.Httpsrv.ReadTimeout,
		"timeout for reading the entire request, including the body")

	flag.DurationVar(&config.Httpsrv.ReadHeaderTimeout, "http-read-header-timeout", config.Httpsrv.ReadHeaderTimeout,
		"timeout for reading the request headers")

	flag.DurationVar(&config.Httpsrv.WriteTimeout, "http-write-timeout", config.Httpsrv.WriteTimeout,
		"timeout for writing the response")

	flag.DurationVar(&config.Httpsrv.IdleTimeout, "http-idle-timeout", config.Httpsrv.IdleTimeout,
		"the maximum amount of time to wait for the next request when keep-alives are enabled")

	flag.DurationVar(&config.Httpsrv.ShutdownTimeout, "http-shutdown-timeout", config.Httpsrv.ShutdownTimeout,
		"timeout for graceful shutdown of the server")

	flag.IntVar(&config.Httpsrv.MaxHeaderBytes, "http-max-header-bytes", config.Httpsrv.MaxHeaderBytes,
		"the maximum number of bytes the server will read parsing the request headers, including the request line")

	// httphandler config

	flag.DurationVar(&config.Httphandler.CacheMaxAge, "http-cache-max-age", config.Httphandler.CacheMaxAge,
		"set max-age in the cache-control header in seconds, set to 0 to disable caching")

	// other config

	flag.BoolVar(&config.PrintConfig, "print-config", config.PrintConfig,
		"print the configuration and exit")

	flag.Parse()

	if err := config.setFromEnv(); err != nil {
		return config, err
	}

	return config, nil
}

func (c Config) setFromEnv() error {
	var err error

	// make a list of flags provided in command line
	provided := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
	})

	flag.VisitAll(func(f *flag.Flag) {
		if err != nil {
			return
		}

		varName := envVarName(f.Name)

		if val, ok := os.LookupEnv(varName); ok {
			if provided[f.Name] {
				err = fmt.Errorf("flag %s is provided in both command line and environment variable", f.Name)
				return
			}

			if errSet := f.Value.Set(val); errSet != nil {
				err = fmt.Errorf("invalid value for a flag in the environment variable %s: %w", varName, errSet)
				return
			}
		}
	})

	return err
}

func envVarName(flagName string) string {
	return envVarPrefix + strings.ReplaceAll(strings.ToUpper(flagName), "-", "_")
}
