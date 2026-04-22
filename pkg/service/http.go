package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"go.uber.org/zap"
)

type HTTPServer struct {
	Server          *http.Server
	shutdownTimeout time.Duration
	done            chan struct{}
	logger          log.Logger
}

func NewHTTPServer(addr string, shutdownTimeout time.Duration, router http.Handler, logger log.Logger) *HTTPServer {
	mux := http.NewServeMux()
	mux.Handle("/", router)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
	}

	return &HTTPServer{
		shutdownTimeout: shutdownTimeout,
		Server:          srv,
		done:            make(chan struct{}),
		logger:          logger.With(zap.String("address", srv.Addr)),
	}
}

func (s *HTTPServer) Start() {
	go func() {
		defer close(s.done)
		s.logger.Info("starting HTTP server")
		if err := s.Server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.With(zap.Error(err)).Fatal("failed to start HTTP server")
		}

		s.logger.Info("http server stopped gracefully")
	}()
}

func (s *HTTPServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	if err := s.Server.Shutdown(ctx); err != nil {
		s.logger.With(zap.Error(err)).Error("http shutdown error")
	}

	s.logger.Info("http server shutdown initiated")
	<-s.done
	cancel()
}
