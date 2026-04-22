package model

import (
	"github.com/gauffer/ddd-events-codegen-converters/pb/identity"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserCreatedEvent — алиас proto-события. Метаданные в events_ext.go
type UserCreatedEvent = identity.UserCreatedEvent

func NewUserCreatedEvent(userID UserID, phone Phone) *UserCreatedEvent {
	return &identity.UserCreatedEvent{
		EventId:    uuid.NewString(),
		UserId:     string(userID),
		Phone:      string(phone),
		OccurredAt: timestamppb.Now(),
	}
}


