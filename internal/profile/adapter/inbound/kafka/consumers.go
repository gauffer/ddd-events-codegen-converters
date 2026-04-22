package kafka

import (
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/application/command"
	"github.com/gauffer/ddd-events-codegen-converters/pb/identity"
	pkgkafka "github.com/gauffer/ddd-events-codegen-converters/pkg/kafka"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"context"
)

const (
	TopicUserCreated = "identity.user_created_event"
)

type Consumers struct {
	UserCreated *pkgkafka.Consumer[identity.UserCreatedEvent]
}

type ConsumersConfig struct {
	ReaderFactory func(ctx context.Context, topic string) pkgkafka.Reader
}

func NewConsumers(
	cfg ConsumersConfig,
	createProfileHandler *command.CreateProfileFromUserEventCommandHandler,
	logger log.Logger,
) *Consumers {
	handler := NewUserCreatedHandler(createProfileHandler, logger)

	return &Consumers{
		UserCreated: pkgkafka.NewConsumer(
			cfg.ReaderFactory,
			TopicUserCreated,
			handler.Handle,
			UnmarshalUserCreatedEvent,
			logger,
		),
	}
}
