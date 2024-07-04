package httpsrv

import (
	"context"
	"fmt"
	"net/http"
)

func NewTLS(ctx context.Context, config ConfigTLS, handler http.Handler) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	srv := &Server{
		config: config.Config,
		tls:    true,
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

	srv.listenAndServeFunc = func() error {
		return srv.server.ListenAndServeTLS(config.CertFile, config.KeyFile)
	}

	return srv, nil
}
