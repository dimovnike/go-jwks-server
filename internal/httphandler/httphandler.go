package httphandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-jwks-server/internal/keyloader"
	"net/http"
	"sync"
	"time"
)

func Handler(kl *keyloader.Keyloader, config Config) http.Handler {
	getKeysJson := getKeysJsonCached(kl)

	apiHandler := func(w http.ResponseWriter) (bool, error) {
		keysJson, cached, err := getKeysJson(kl.GetKeysLoadTime())
		if err != nil {
			return cached, fmt.Errorf("getting keys: %w", err)
		}

		if config.CacheMaxAge > 0 {
			w.Header().Set("cache-control", fmt.Sprintf("max-age=%d", int(config.CacheMaxAge.Seconds())))
		}

		w.Write(keysJson)

		return cached, nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.With().Str("host", r.Host).Str("uri", r.RequestURI).Str("remote_address", r.RemoteAddr).Logger()

		w.Header().Set("Content-Type", "application/json")

		if code, ok := httpValidateRequest(r); !ok {
			errText := http.StatusText(code)
			logger.Error().Err(errors.New(errText)).Int("code", code).Msg("failed to validate request")
			http.Error(w, `"`+errText+`"`, code) // text in JSON format
			return
		}

		cached, err := apiHandler(w)
		if err != nil {
			logger.Error().Err(err).Bool("cached", cached).Msg("failed to process request")
			http.Error(w, `"internal server error"`, http.StatusInternalServerError)
			return
		}

		logger.Info().Bool("cached", cached).Msg("request processed successfully")
	})
}

func GetKeysJson(kl *keyloader.Keyloader) (keysJson []byte, keysLoadTime time.Time, _err error) {
	keys, loadTime, err := kl.GetKeys()
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("getting keys: %w", err)
	}

	j, err := json.Marshal(keys)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("marshalling keys: %w", err)
	}

	return j, loadTime, nil
}

// getKeysJsonCached cached GetKeysJson function by keysLoadTime
// the returned function is safe for concurrent use
func getKeysJsonCached(kl *keyloader.Keyloader) func(time.Time) ([]byte, bool, error) {
	var lastKeysLoadTime time.Time = time.Date(0, 0, 0, 0, 0, 0, 1, time.UTC) // trigger load on first call
	var lastKeysJson []byte
	var lastErr error = errors.New("no keys loaded")

	var m sync.RWMutex

	return func(loadTime time.Time) ([]byte, bool, error) {
		m.RLock()
		mustLoad := !lastKeysLoadTime.Equal(loadTime)
		m.RUnlock()

		if !mustLoad {
			return lastKeysJson, true, lastErr
		}

		keys, loadTime, err := GetKeysJson(kl)

		m.Lock()
		lastKeysJson = keys
		lastKeysLoadTime = loadTime
		lastErr = err
		m.Unlock()

		return keys, false, err
	}
}

func httpValidateRequest(r *http.Request) (int, bool) {
	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, false
	}

	if r.URL.Path != "/keys" {
		return http.StatusNotFound, false
	}

	return 0, true
}
