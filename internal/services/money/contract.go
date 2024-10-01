package moneyservice

import (
	"context"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/helpers/transaction"
	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
)

type WithdrawalStorage interface {
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]withdrawal.Withdrawal, error)
	GetWithdrawal(ctx context.Context, ID uuid.UUID) (*withdrawal.Withdrawal, error)
	InsertWithdrawal(ctx context.Context, inputWithdrawal *withdrawal.Withdrawal, trx *transaction.Trx) error
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

type BalanceStorage interface {
	AddBalance(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error
	ReduceBalance(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error
	GetBalance(ctx context.Context, userID uuid.UUID) (*balance.Balance, error)
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}
