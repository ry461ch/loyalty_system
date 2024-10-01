package moneyhandlers

import (
	"context"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
)

type MoneyService interface {
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]withdrawal.Withdrawal, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (*balance.Balance, error)
	Withdraw(ctx context.Context, withdrawal *withdrawal.Withdrawal) error
}
