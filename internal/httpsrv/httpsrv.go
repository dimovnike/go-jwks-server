package httpsrv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

type Server struct {
	config Config
	server *http.Server
}

func New(ctx context.Context, config Config, handler http.Handler) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	srv := &Server{
		config: config,
		server: &http.Server{
			Addr:              config.Addr,
			ReadTimeout:       config.ReadTimeout,
			ReadHeaderTimeout: config.ReadHeaderTimeout,
			WriteTimeout:      config.WriteTimeout,
			IdleTimeout:       config.IdleTimeout,
			MaxHeaderBytes:    config.MaxHeaderBytes,

			Handler: handler,
		},
	}

	return srv, nil
}

func (s *Server) ListenAndServeWithCtx(ctx context.Context) error {
	wg := sync.WaitGroup{}

	var serveErr error

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := s.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			serveErr = err
		}

		cancel()
		log.Debug().Err(err).Msg("HTTP server goroutine exited")
	}()

	log.Info().Msg("HTTP server started")
	defer log.Info().Msg("HTTP server stopped")

	<-ctx.Done()

	log.Info().Msg("HTTP server stopping ...")

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer shutdownCtxCancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}

	wg.Wait()

	return serveErr
}
