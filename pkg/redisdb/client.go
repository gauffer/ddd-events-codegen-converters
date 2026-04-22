package redisdb

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/service"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var _ service.StartStopper = (*Client)(nil)

type Client struct {
	*redis.Client

	logger log.Logger
}

func NewClient(uri string, logger log.Logger) *Client {
	opts, err := redis.ParseURL(uri)
	if err != nil {
		// Fallback: URI не в формате redis://
		opts = &redis.Options{
			Addr: uri,
		}
	}

	return &Client{
		Client: redis.NewClient(opts),
		logger: logger.With(zap.String("component", "redis")),
	}
}

func (c *Client) Start() {
	if err := c.Client.Ping(context.Background()).Err(); err != nil {
		c.logger.With(zap.Error(err)).Fatal("failed to ping Redis")
	}

	c.logger.Info("Redis pinged successfully")
}

func (c *Client) Stop() {
	if err := c.Client.Close(); err != nil {
		c.logger.With(zap.Error(err)).Error("failed to close Redis")
	}

	c.logger.Info("Redis closed successfully")
}
