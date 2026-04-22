package command

import (
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/domain/repository"

	"context"
	"fmt"
)

type CreateProfileFromUserEventCommand struct {
	UserID string
}

type CreateProfileFromUserEventCommandHandler struct {
	profiles repository.ProfilesRepository
}

func NewCreateProfileFromUserEventCommandHandler(
	profiles repository.ProfilesRepository,
) *CreateProfileFromUserEventCommandHandler {
	return &CreateProfileFromUserEventCommandHandler{
		profiles: profiles,
	}
}

func (h *CreateProfileFromUserEventCommandHandler) Handle(
	ctx context.Context,
	cmd CreateProfileFromUserEventCommand,
) error {
	existing, err := h.profiles.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("find profile by user ID: %w", err)
	}

	if existing != nil {
		return nil
	}

	profile, err := model.NewProfile(cmd.UserID)
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}

	if saveErr := h.profiles.Save(ctx, profile); saveErr != nil {
		return fmt.Errorf("save profile: %w", saveErr)
	}

	return nil
}
