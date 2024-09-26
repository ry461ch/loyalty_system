package userservice

import (
	"context"

	"github.com/ry461ch/loyalty_system/internal/helpers/transaction"
	"github.com/ry461ch/loyalty_system/internal/models/user"
)

type UserStorage interface {
	GetUser(ctx context.Context, login string) (*user.User, error)
	InsertUser(ctx context.Context, newUser *user.User, trx *transaction.Trx) error
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}
