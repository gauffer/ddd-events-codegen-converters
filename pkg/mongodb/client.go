package mongodb

import (
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/service"

	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var _ service.StartStopper = (*Client)(nil)

const defaultTimeout = 10 * time.Second

type Client struct {
	*mongo.Client

	uri    string
	logger log.Logger
}

func NewClient(uri string, logger log.Logger) *Client {
	return &Client{
		uri:    uri,
		logger: logger.With(zap.String("component", "mongodb")),
	}
}

func (m *Client) Start() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	m.logger.Info("connecting to MongoDB")

	clientOptions := options.Client().ApplyURI(m.uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		m.logger.With(zap.Error(err)).Fatal("failed to connect to MongoDB")
	}

	if pingErr := client.Ping(ctx, nil); pingErr != nil {
		m.logger.With(zap.Error(pingErr)).Fatal("failed to ping MongoDB")
	}

	m.Client = client

	m.logger.Info("MongoDB connected successfully")
}

func (m *Client) Stop() {
	if m.Client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	m.logger.Info("disconnecting from MongoDB")

	if err := m.Client.Disconnect(ctx); err != nil {
		m.logger.With(zap.Error(err)).Error("failed to disconnect from MongoDB")
		return
	}

	m.Client = nil

	m.logger.Info("MongoDB disconnected successfully")
}

func (m *Client) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (any, error)) error {
	session, err := m.StartSession()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, fn)
	if err != nil {
		return err
	}

	return nil
}
