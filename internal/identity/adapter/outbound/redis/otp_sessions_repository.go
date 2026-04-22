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

var _ repository.OTPSessionsRepository = (*OTPSessionsRepository)(nil)

type OTPSessionsRepository struct {
	client *redis.Client
}

func NewOTPSessionsRepository(client *redis.Client) repository.OTPSessionsRepository {
	return &OTPSessionsRepository{
		client: client,
	}
}

func (r *OTPSessionsRepository) Save(ctx context.Context, session *model.OTPSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal otp session: %w", err)
	}

	key := r.key(session.ID)
	phoneIndexKey := r.phoneIndexKey(session.Phone)
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.Set(ctx, phoneIndexKey, string(session.ID), ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("save otp session to redis: %w", err)
	}

	return nil
}

func (r *OTPSessionsRepository) FindByID(ctx context.Context, id model.OTPSessionID) (*model.OTPSession, error) {
	key := r.key(id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, model.ErrPhoneAuthSessionNotFound
		}
		return nil, fmt.Errorf("get otp session from redis: %w", err)
	}

	var session model.OTPSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal otp session: %w", err)
	}

	return &session, nil
}

func (r *OTPSessionsRepository) FindActiveByPhone(ctx context.Context, phone model.Phone) (*model.OTPSession, error) {
	phoneIndexKey := r.phoneIndexKey(phone)
	sessionIDStr, err := r.client.Get(ctx, phoneIndexKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("get otp session id by phone from redis: %w", err)
	}

	sessionID := model.OTPSessionID(sessionIDStr)
	return r.FindByID(ctx, sessionID)
}

func (r *OTPSessionsRepository) Delete(ctx context.Context, id model.OTPSessionID) error {
	// Получаем сессию для удаления индекса по телефону
	session, err := r.FindByID(ctx, id)
	if err != nil && err != model.ErrPhoneAuthSessionNotFound {
		return fmt.Errorf("find otp session before delete: %w", err)
	}

	key := r.key(id)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)

	if session != nil {
		phoneIndexKey := r.phoneIndexKey(session.Phone)
		pipe.Del(ctx, phoneIndexKey)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("delete otp session from redis: %w", err)
	}

	return nil
}

func (r *OTPSessionsRepository) key(id model.OTPSessionID) string {
	return fmt.Sprintf("otp_session:%s", string(id))
}

func (r *OTPSessionsRepository) phoneIndexKey(phone model.Phone) string {
	return fmt.Sprintf("otp_session:phone:%s", string(phone))
}
