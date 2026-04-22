package app

import (
	"context"
	"strings"

	"github.com/gauffer/ddd-events-codegen-converters/internal/identity"
	identityFake "github.com/gauffer/ddd-events-codegen-converters/internal/identity/adapter/outbound/fake"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/adapter/outbound/crypto"
	"github.com/gauffer/ddd-events-codegen-converters/internal/pkg/jws"
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile"
	pkgkafka "github.com/gauffer/ddd-events-codegen-converters/pkg/kafka"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/outbox"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/redisdb"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/service"
)

func Run() {
	cfg := NewConfig()
	logger := log.NewLogger()

	mongoClient := mongodb.NewClient(cfg.MongoURI, logger)
	redisClient := redisdb.NewClient(cfg.RedisURI, logger)
	outboxPublisher := outbox.NewPublisher(mongoClient, "app", "identity_outbox")

	jwtSigner, err := crypto.NewJWTSigner(cfg.JWTPrivateKeyB64, "demo")
	if err != nil {
		logger.Fatal("failed to create JWT signer")
	}
	jwtValidator, err := jws.NewValidator(jwtSigner.GetPublicKeyPEM(), "demo", "")
	if err != nil {
		logger.Fatal("failed to create JWT validator")
	}

	identityAdapters := identity.NewAdapters(
		service.Environment(cfg.AppEnv),
		logger,
		identity.WithMongoClient(mongoClient),
		identity.WithRedisClient(redisClient.Client),
		identity.WithOutboxPublisher(outboxPublisher),
		identity.WithOTPSender(identityFake.NewOTPSender(logger)),
		identity.WithJWTSigner(jwtSigner),
	)
	identityModule := identity.NewModule(logger, identityAdapters)

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	kafkaReaderFactory := pkgkafka.NewReaderFactory(
		brokers, cfg.KafkaSASLUser, cfg.KafkaSASLPassword,
		cfg.KafkaSASLEnabled, 1, 10e6, "demo-api", logger,
	)

	profileAdapters := profile.NewAdapters(profile.WithMongoClient(mongoClient))
	profileModule := profile.NewModule(
		logger, profileAdapters,
		profile.WithKafkaReaderFactory(func(ctx context.Context, topic string) pkgkafka.Reader {
			return kafkaReaderFactory(ctx, topic)
		}),
	)

	httpServer := newServer(
		cfg.RestAddress, cfg.ShutdownTimeout, logger,
		identityModule.GetHTTPHandlers(), jwtValidator, mongoClient, redisClient.Client,
	)

	userCreatedWorker := service.NewWorker(
		"user-created-consumer",
		profileModule.GetKafkaConsumers().UserCreated,
		logger,
	)

	container := service.NewContainer(
		service.WithLogger(logger),
		service.WithAppName("demo-api"),
		service.WithEnvironment(cfg.AppEnv),
		service.WithDiagnosticsServer(cfg.DiagnosticsAddress),
	)

	container.RunWait(mongoClient, redisClient, httpServer, userCreatedWorker).Wait()
}
