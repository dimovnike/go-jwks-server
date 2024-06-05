package main

import (
	"context"
	"flag"
	"fmt"
	"go-jwks-server/internal/config"
	"go-jwks-server/internal/httphandler"
	"go-jwks-server/internal/httpsrv"
	"go-jwks-server/internal/keyfiles"
	"go-jwks-server/internal/keyloader"
	"go-jwks-server/internal/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

var log zerolog.Logger

func main() {
	config, err := config.New()
	if err != nil {
		fmt.Println("failed to create config:", err)
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(1)
	}

	log, err = logger.New(config.Logger)
	if err != nil {
		fmt.Println("failed to create logger:", err)
		os.Exit(1)
	}

	if config.PrintConfig {
		fmt.Println(config.String())
		os.Exit(0)
	}

	keyfiles.SetLogger(log)
	keyloader.SetLogger(log)
	httpsrv.SetLogger(log)
	httphandler.SetLogger(log)

	ctx, cancel := shutdownContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	kl, err := keyloader.NewKeyloader(config.Keyloader)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create keyloader")
		return
	}

	srv, err := httpsrv.New(ctx, config.Httpsrv, httphandler.Handler(kl, config.Httphandler))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create http srv")
		return
	}

	if !config.Keyloader.WatchOn() {
		if err := kl.LoadKeys(); err != nil {
			log.Fatal().Err(err).Msg("failed to load keys")
			return
		}
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := srv.ListenAndServeWithCtx(ctx); err != nil {
			return fmt.Errorf("http server: %w", err)
		}

		return nil
	})

	if config.Keyloader.WatchOn() {
		eg.Go(func() error {
			if err := kl.LoadKeysWatch(ctx); err != nil {
				return fmt.Errorf("keyloader: %w", err)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("service terminated with error")
		return
	}

	log.Info().Msg("service terminated")
}

// shutdownContext replicates signal.NotifyContext but also logs the received signal
func shutdownContext(ctx context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	if ctx.Err() == nil {
		go func() {
			select {
			case sig := <-ch:
				log.Info().Str("signal", sig.String()).Msg("received signal")
				cancel()
			case <-ctx.Done():
			}
		}()
	}
	return ctx, cancel
}
