package httpsrv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
)

type Server struct {
	config             Config
	server             *http.Server
	tls                bool
	listenAndServeFunc func() error
	log                zerolog.Logger
}

func New(ctx context.Context, config Config, handler http.Handler) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	httpSrv := &http.Server{
		Addr:              config.Addr,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,

		Handler: handler,
	}

	srv := &Server{
		config:             config,
		server:             httpSrv,
		listenAndServeFunc: httpSrv.ListenAndServe,
	}

	return srv, nil
}

func (s *Server) SetLogger(logger zerolog.Logger) {
	s.log = logger.With().Bool("https", s.tls).Logger()
}

func (s *Server) ListenAndServeWithCtx(ctx context.Context) error {
	wg := sync.WaitGroup{}

	var serveErr error

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := s.listenAndServeFunc()
		if !errors.Is(err, http.ErrServerClosed) {
			serveErr = err
		}

		cancel()
		s.log.Debug().Err(err).Msg("HTTP server goroutine exited")
	}()

	logger := s.log.With().Str("listen", s.config.Addr).Logger()

	logger.Info().Msg("HTTP server started")
	defer logger.Info().Msg("HTTP server stopped")

	<-ctx.Done()

	logger.Info().Msg("HTTP server stopping ...")

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer shutdownCtxCancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}

	wg.Wait()

	return serveErr
}
