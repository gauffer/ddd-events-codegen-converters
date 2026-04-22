// Package service — graceful shutdown и диагностика
package service

import (
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"os"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Environment string

const (
	EnvLocal      Environment = "local"
	EnvProduction Environment = "production"
)

type Container struct {
	logger             log.Logger
	status             *Status
	diagnosticsService *DiagnosticsService
	environment        Environment

	appName string
}

type StartStopper interface {
	Start()
	Stop()
}

func NewContainer(o ...Option) *Container {
	var opts options

	for _, opt := range o {
		opt(&opts)
	}

	if err := opts.Validate(); err != nil {
		panic(err)
	}

	c := &Container{
		logger:      opts.Logger,
		appName:     opts.AppName,
		environment: opts.Environment,

		status: &Status{
			logger: opts.Logger,
		},
	}

	if opts.DiagnosticsServerAddr != "" {
		c.diagnosticsService = NewDiagnosticsService(
			opts.DiagnosticsServerAddr,
			c.status,
			c.logger,
			time.Second,
		)
	}

	// @TODO: настроить labels

	c.logger.Info("initializing app")
	c.status.SetAlive(true)

	return c
}

func (c *Container) RunWait(services ...StartStopper) *sync.WaitGroup {
	c.logger.Info("starting app")

	if c.diagnosticsService != nil {
		c.diagnosticsService.Start()
	}

	for _, s := range services {
		s.Start()
	}

	c.status.SetReady(true)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		c.logger.With(
			zap.String("signal", Wait([]os.Signal{syscall.SIGTERM, syscall.SIGINT}).String()),
		).Info("received signal. stopping")

		c.status.SetReady(false)
		for i := range services {
			services[len(services)-i-1].Stop()
		}

		if c.diagnosticsService != nil {
			c.diagnosticsService.Stop()
		}

		c.logger.Info("bye")
	}()

	return wg
}

