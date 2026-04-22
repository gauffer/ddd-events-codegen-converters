package redis

import (
	"context"
	"encoding/json"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ repository.AuthorizationCodesRepository = (*AuthCodesRepository)(nil)

type AuthCodesRepository struct {
	client *redis.Client
}

func NewAuthCodesRepository(client *redis.Client) repository.AuthorizationCodesRepository {
	return &AuthCodesRepository{
		client: client,
	}
}

func (r *AuthCodesRepository) Save(ctx context.Context, code *model.AuthorizationCode) error {
	data, err := json.Marshal(code)
	if err != nil {
		return fmt.Errorf("marshal authorization code: %w", err)
	}

	key := r.key(code.ID)
	ttl := time.Until(code.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("save authorization code to redis: %w", err)
	}

	return nil
}

func (r *AuthCodesRepository) GetByID(ctx context.Context, id model.AuthorizationCodeID) (*model.AuthorizationCode, error) {
	key := r.key(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, model.ErrAuthorizationCodeNotFound
		}
		return nil, fmt.Errorf("get authorization code from redis: %w", err)
	}

	var code model.AuthorizationCode
	if err := json.Unmarshal(data, &code); err != nil {
		return nil, fmt.Errorf("unmarshal authorization code: %w", err)
	}

	return &code, nil
}

func (r *AuthCodesRepository) Delete(ctx context.Context, id model.AuthorizationCodeID) error {
	key := r.key(id)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete authorization code from redis: %w", err)
	}

	return nil
}

func (r *AuthCodesRepository) key(id model.AuthorizationCodeID) string {
	return fmt.Sprintf("auth_code:%s", string(id))
}
