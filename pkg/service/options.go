package service

import (
	"errors"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
)

type options struct {
	DiagnosticsServerAddr string
	Logger                log.Logger
	AppName               string
	Environment           Environment
}

type Option func(o *options)

func WithDiagnosticsServer(addr string) Option {
	return func(o *options) {
		o.DiagnosticsServerAddr = addr
	}
}

func WithLogger(log log.Logger) Option {
	return func(o *options) {
		o.Logger = log
	}
}

func WithAppName(name string) Option {
	return func(o *options) {
		o.AppName = name
	}
}

func WithEnvironment(env string) Option {
	return func(o *options) {
		switch env {
		case "local":
			o.Environment = EnvLocal
		case "production":
			o.Environment = EnvProduction
		}
	}
}

func (o *options) Validate() error {
	if o.Logger == nil {
		o.Logger = log.NewLogger()
	}

	if o.Environment == "" {
		return errors.New("environment is required")
	}

	return nil
}
