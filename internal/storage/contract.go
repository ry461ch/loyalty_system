package storage

import (
	"context"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type UserStorage interface {
	GetUser(ctx context.Context, login string) (*user.User, error)
	InsertUser(ctx context.Context, newUser *user.User, trx *transaction.Trx) error
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

type OrderStorage interface {
	GetOrderUserID(ctx context.Context, orderID string) (*uuid.UUID, error)
	InsertOrder(ctx context.Context, userID uuid.UUID, orderID string, trx *transaction.Trx) error
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error)
	GetWaitingOrderIDs(ctx context.Context, limit, offset int) ([]string, error)
	UpdateOrder(ctx context.Context, order *order.Order, tx *transaction.Trx) (*uuid.UUID, error)
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

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
