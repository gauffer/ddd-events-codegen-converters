package kafka

import (
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.uber.org/zap"
)

const (
	defaultMaxWait       = 500 * time.Millisecond
	defaultDialerTimeout = 10 * time.Second
	defaultSleepDuration = 100 * time.Millisecond
)

type Reader interface {
	FetchMessage(ctx context.Context) (RawMessage, error)
	CommitMessages(ctx context.Context, msgs ...RawMessage) error
	Close() error
}

func NewReaderFactory(
	brokers []string,
	saslUser string,
	saslPassword string,
	saslEnabled bool,
	minFetchBytes int,
	maxFetchBytes int,
	groupBase string,
	logger log.Logger,
	postfixes ...string,
) func(ctx context.Context, topic string) Reader {
	parts := append([]string{groupBase}, postfixes...)
	groupID := strings.Join(parts, ".")
	sugar := logger.Zap().Sugar()

	return func(_ context.Context, topic string) Reader {
		readerConfig := kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			GroupTopics:    []string{topic},
			MinBytes:       minFetchBytes,
			MaxBytes:       maxFetchBytes,
			MaxWait:        defaultMaxWait,
			CommitInterval: time.Second,
			StartOffset:    kafka.FirstOffset,
			ErrorLogger: kafka.LoggerFunc(
				func(format string, msgArgs ...any) {
					sugar.Errorf(format, msgArgs...)
				},
			),
			Logger: kafka.LoggerFunc(
				func(format string, msgArgs ...any) {
					sugar.Debugf(format, msgArgs...)
				},
			),
			WatchPartitionChanges: true,
		}

		if saslEnabled {
			mechanism, err := scram.Mechanism(scram.SHA256, saslUser, saslPassword)
			if err != nil {
				sugar.Errorf("failed to create SCRAM mechanism: %v", err)
			} else {
				readerConfig.Dialer = &kafka.Dialer{
					Timeout:       defaultDialerTimeout,
					DualStack:     true,
					SASLMechanism: mechanism,
				}
			}
		}

		return kafka.NewReader(readerConfig)
	}
}

type Consumer[T any] struct {
	topic     string
	newReader func(ctx context.Context, topic string) Reader
	handler   func(ctx context.Context, event *T) error
	decoder   func([]byte, *T) error
	logger    log.Logger
}

func NewConsumer[T any](
	factory func(ctx context.Context, topic string) Reader,
	topic string,
	handler func(ctx context.Context, event *T) error,
	decoder func([]byte, *T) error,
	logger log.Logger,
) *Consumer[T] {
	return &Consumer[T]{
		topic:     topic,
		newReader: factory,
		handler:   handler,
		decoder:   decoder,
		logger:    logger,
	}
}

func (c *Consumer[T]) Run(ctx context.Context) error {
	reader := c.newReader(ctx, c.topic)
	defer func() {
		if err := reader.Close(); err != nil {
			c.logger.Error("reader close", zap.Error(err))
		}
	}()

	for ctx.Err() == nil {
		msg, fetchErr := reader.FetchMessage(ctx)
		switch {
		case fetchErr == nil:
		case errors.Is(fetchErr, context.DeadlineExceeded), errors.Is(fetchErr, context.Canceled):
			c.logger.Info("stop consumer")
			return nil
		case errors.Is(fetchErr, io.EOF):
			c.logger.Warn("fetch message return EOF")
			reader = c.newReader(ctx, c.topic)
			c.logger.Info("recreating reader")
			continue
		default:
			c.logger.Error("fetch message", zap.Error(fetchErr))
			c.asyncSleep(defaultSleepDuration)
			reader = c.newReader(ctx, c.topic)
			continue
		}

		var event T
		if decodeErr := c.decoder(msg.Value, &event); decodeErr != nil {
			c.logger.Error("decode message",
				zap.String("topic", msg.Topic),
				zap.String("key", string(msg.Key)),
				zap.Error(decodeErr))
			c.asyncSleep(defaultSleepDuration)
			reader = c.newReader(ctx, c.topic)
			continue
		}

		if handleErr := c.handler(ctx, &event); handleErr != nil {
			c.logger.Error("handler execution",
				zap.String("topic", msg.Topic),
				zap.String("key", string(msg.Key)),
				zap.Error(handleErr))
			c.asyncSleep(defaultSleepDuration)
			reader = c.newReader(ctx, c.topic)
			continue
		}

		if commitErr := reader.CommitMessages(context.WithoutCancel(ctx), msg); commitErr != nil {
			c.logger.Error("commit message", zap.Error(commitErr))
		}
	}

	return nil
}

func (c *Consumer[T]) asyncSleep(duration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	<-ctx.Done()
}
