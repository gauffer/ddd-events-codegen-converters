package service

import (
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"context"

	"go.uber.org/zap"
)

type runner interface {
	Run(ctx context.Context) error
}

type Worker struct {
	runner runner
	logger log.Logger
	name   string
	cancel context.CancelFunc
	done   chan struct{}
}

func NewWorker(name string, runner runner, logger log.Logger) *Worker {
	return &Worker{
		runner: runner,
		logger: logger.With(zap.String("worker", name)),
		name:   name,
		done:   make(chan struct{}),
	}
}

func (w *Worker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go func() {
		defer close(w.done)
		w.logger.Info("starting worker")

		if err := w.runner.Run(ctx); err != nil && ctx.Err() == nil {
			w.logger.Error("worker error", zap.Error(err))
		}

		w.logger.Info("worker stopped")
	}()
}

func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}

	<-w.done
}

var _ StartStopper = (*Worker)(nil)
