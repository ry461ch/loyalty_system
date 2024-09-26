package authhandlers

import (
	"context"

	"github.com/ry461ch/loyalty_system/internal/models/user"
)

type UserService interface {
	Login(ctx context.Context, inputUser *user.InputUser) (*string, error)
	Register(ctx context.Context, inputUser *user.InputUser) (*string, error)
}
