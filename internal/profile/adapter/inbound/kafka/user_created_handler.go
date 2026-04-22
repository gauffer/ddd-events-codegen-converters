package kafka

import (
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/application/command"
	"github.com/gauffer/ddd-events-codegen-converters/pb/identity"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"context"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type UserCreatedHandler struct {
	createProfileHandler *command.CreateProfileFromUserEventCommandHandler
	logger               log.Logger
}

func NewUserCreatedHandler(
	createProfileHandler *command.CreateProfileFromUserEventCommandHandler,
	logger log.Logger,
) *UserCreatedHandler {
	return &UserCreatedHandler{
		createProfileHandler: createProfileHandler,
		logger:               logger,
	}
}

func (h *UserCreatedHandler) Handle(ctx context.Context, event *identity.UserCreatedEvent) error {
	err := h.createProfileHandler.Handle(ctx, command.CreateProfileFromUserEventCommand{
		UserID: event.GetUserId(),
	})
	if err != nil {
		return err
	}

	h.logger.Info("profile created from user_created event",
		zap.String("user_id", event.GetUserId()))

	return nil
}

func UnmarshalUserCreatedEvent(data []byte, event *identity.UserCreatedEvent) error {
	return proto.Unmarshal(data, event)
}
