package profile

import (
	"context"

	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/adapter/inbound/kafka"
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/application/command"
	pkgkafka "github.com/gauffer/ddd-events-codegen-converters/pkg/kafka"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
)

type Module struct {
	kafkaConsumers *kafka.Consumers
}

type ModuleOptions func(*Module, *Adapters, log.Logger)

func WithKafkaReaderFactory(factory func(ctx context.Context, topic string) pkgkafka.Reader) ModuleOptions {
	return func(m *Module, adapters *Adapters, logger log.Logger) {
		createProfileHandler := command.NewCreateProfileFromUserEventCommandHandler(
			adapters.GetProfilesRepository(),
		)

		m.kafkaConsumers = kafka.NewConsumers(
			kafka.ConsumersConfig{ReaderFactory: factory},
			createProfileHandler,
			logger,
		)
	}
}

func NewModule(
	logger log.Logger,
	adapters *Adapters,
	options ...ModuleOptions,
) *Module {
	m := &Module{}

	for _, opt := range options {
		opt(m, adapters, logger)
	}

	return m
}

func (m *Module) GetKafkaConsumers() *kafka.Consumers {
	return m.kafkaConsumers
}
