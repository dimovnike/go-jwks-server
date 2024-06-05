package keyloader

import (
	"context"
	"errors"
	"go-jwks-server/internal/keyfiles"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
)

/*
	this package watches a directory for changes and loads public keys from files in that directory
	file names must be the key name and the file content must be the key value

	key Id is derived from the file name, the .pub extension is removed if present
	to ignore a file, add a .ignore extension
*/

type Keyloader struct {
	config Config

	// the keys loaded from the directory
	keys              jwk.Set
	keysLoadTimestamp time.Time

	// the mutex to protect the keys and keysTimestamp
	m sync.RWMutex
}

func NewKeyloader(config Config) (*Keyloader, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	kl := &Keyloader{
		config: config,
	}

	return kl, nil
}

func (kl *Keyloader) GetKeysLoadTime() time.Time {
	kl.m.RLock()
	defer kl.m.RUnlock()

	return kl.keysLoadTimestamp
}

// GetKeys returns a copy of the keys
func (kl *Keyloader) GetKeys() (jwk.Set, time.Time, error) {
	kl.m.RLock()
	defer kl.m.RUnlock()

	if kl.keys == nil {
		return nil, time.Time{}, errors.New("keys not loaded")
	}

	return kl.keys, kl.keysLoadTimestamp, nil
}

// LoadKeysWatch starts watching the directory for changes and loads the keys
// it honors the FailOnError config option
func (kl *Keyloader) LoadKeysWatch(ctx context.Context) error {
	watcher := keyfiles.NewWatcher()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := watcher.Watch(ctx, kl.config.Dir, kl.config.WatchInterval)
		log.Debug().Err(err).Msg("watcher goroutine exited")

		cancel()
	}()

	log.Info().Str("dir", kl.config.Dir).Dur("interval", kl.config.WatchInterval).Msg("started watching directory for changes")
	defer log.Info().Msg("stopped watching directory for changes")

	var retErr error

	// watcher will close the channel when done
	for event := range watcher.Events {
		if event.Error != nil {
			if kl.config.FailOnError {
				retErr = event.Error
				cancel()
				break
			}

			log.Error().Err(event.Error).Msg("watcher event error")
			continue
		}

		if err := kl.LoadKeys(); err != nil {
			retErr = err
			cancel()
			break
		}
	}

	log.Info().Str("dir", kl.config.Dir).Msg("stopping watching directory for changes ...")

	wg.Wait()

	return retErr
}

// LoadKeysOnce loads the keys once
// it honors the FailOnError config option
func (kl *Keyloader) LoadKeys() error {
	keys, err := loadKeys(kl.config.Dir)
	if err != nil {
		if kl.config.FailOnError {
			return err
		}

		log.Error().Err(err).Msg("failed to load keys")
		return nil // leave the old keys
	}

	kl.m.Lock()
	defer kl.m.Unlock()

	kl.keys = keys
	kl.keysLoadTimestamp = time.Now()

	return nil
}
