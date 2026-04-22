package model

import (
	"fmt"
	"time"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/domain"
	"github.com/google/uuid"
)

type UserID string
type UserStatus string

const (
	UserStatusActive UserStatus = "active"
)

type User struct {
	domain.AggregateRoot

	ID       UserID
	PublicID *int64
	Phone    Phone
	Status   UserStatus

	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
}

func NewUser(phone Phone) (*User, error) {
	if !phone.IsValid() {
		return nil, ErrInvalidPhone
	}

	now := time.Now()
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("generating user uuid: %w", err)
	}

	user := &User{
		ID:        UserID(id.String()),
		Phone:     phone,
		Status:    UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	user.AddEvent(NewUserCreatedEvent(user.ID, user.Phone))

	return user, nil
}

func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

func (u *User) IsBlocked() bool {
	return false
}
