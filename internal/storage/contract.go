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
	GetOrderUserId(ctx context.Context, orderId string) (*uuid.UUID, error)
	InsertOrder(ctx context.Context, userId uuid.UUID, orderId string, trx *transaction.Trx) error
	GetOrders(ctx context.Context, userId uuid.UUID) ([]order.Order, error)
	UpdateOrder(ctx context.Context, order *order.Order, trx *transaction.Trx) error
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

type WithdrawalStorage interface {
	GetWithdrawals(ctx context.Context, userId uuid.UUID) ([]withdrawal.Withdrawal, error)
	GetWithdrawal(ctx context.Context, Id uuid.UUID) (*withdrawal.Withdrawal, error)
	InsertWithdrawl(ctx context.Context, userId uuid.UUID, withdrawal *withdrawal.Withdrawal, trx *transaction.Trx) error
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

type BalanceStorage interface {
	AddBalance(ctx context.Context, userId uuid.UUID, amount float64, trx *transaction.Trx) error
	ReduceBalance(ctx context.Context, userId uuid.UUID, amount float64, trx *transaction.Trx) error
	GetBalance(ctx context.Context, userId uuid.UUID) (*balance.Balance, error)
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

type MoneyStorage interface {
	BalanceStorage
	WithdrawalStorage

	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

type Storage interface {
	MoneyStorage
	OrderStorage
	UserStorage

	Initialize(ctx context.Context) error
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}
