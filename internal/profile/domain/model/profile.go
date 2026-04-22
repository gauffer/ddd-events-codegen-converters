package model

import (
	"fmt"
	"time"
)

type ProfileID string

type Profile struct {
	ID        ProfileID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewProfile(userID string) (*Profile, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	now := time.Now()

	return &Profile{
		ID:        ProfileID(userID),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
