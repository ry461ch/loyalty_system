package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type UserService interface {
	Login(ctx context.Context, inputUser *user.InputUser) (*string, error)
	Register(ctx context.Context, inputUser *user.InputUser) (*string, error)
}

type MoneyService interface {
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]withdrawal.Withdrawal, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (*balance.Balance, error)
	AddAccrual(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error
	Withdraw(ctx context.Context, withdrawal *withdrawal.Withdrawal) error
}

type OrderService interface {
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error)
	InsertOrder(ctx context.Context, userID uuid.UUID, orderID string) error
	UpdateOrder(ctx context.Context, updatedOrder *order.Order) error
}
