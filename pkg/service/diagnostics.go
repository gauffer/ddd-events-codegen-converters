package service

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	DefaultReadHeaderTimeout = 10 * time.Second
)

type DiagnosticsService struct {
	server *http.Server
	logger log.Logger

	shutdownTimeout time.Duration
	done            chan struct{}
}

func NewDiagnosticsService(
	addr string,
	status *Status,
	logger log.Logger,
	shutdownTimeout time.Duration,
) *DiagnosticsService {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		if status.IsAlive() {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
	})

	// @TODO: реализовать
	mux.HandleFunc("/info", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if status.IsReady() {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
	})

	return &DiagnosticsService{
		server: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: DefaultReadHeaderTimeout,
		},
		logger: logger.With(
			zap.String("address", addr),
		),

		shutdownTimeout: shutdownTimeout,
		done:            make(chan struct{}),
	}
}

func (s *DiagnosticsService) Start() {
	go func() {
		defer close(s.done)
		s.logger.Info("starting diagnostics service")

		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.With(zap.Error(err)).Fatal("diagnostics service failure")
		}

		s.logger.Info("diagnostics service stop listening")
	}()
}

func (s *DiagnosticsService) Stop() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.shutdownTimeout,
	)
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.With(zap.Error(err)).Error("http shutdown error")
	}
	s.logger.Info("http server stopped")
	<-s.done
	cancel()
}
