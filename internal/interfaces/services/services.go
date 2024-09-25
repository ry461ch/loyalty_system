package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/helpers/transaction"
	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/models/user"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
)

type UserService interface {
	Login(ctx context.Context, inputUser *user.InputUser) (*string, error)
	Register(ctx context.Context, inputUser *user.InputUser) (*string, error)
}

type WaitingOrdersGetterService interface {
	GetWaitingOrders(ctx context.Context, limit int, createdAt *time.Time) ([]order.Order, error)
}

type OrderUpdaterService interface {
	UpdateOrder(ctx context.Context, inputOrder *order.Order) error
}

type OrderService interface {
	WaitingOrdersGetterService
	OrderUpdaterService

	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error)
	InsertOrder(ctx context.Context, userID uuid.UUID, orderID string) error
}

type AccrualAdderService interface {
	AddAccrual(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error
}

type MoneyService interface {
	AccrualAdderService

	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]withdrawal.Withdrawal, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (*balance.Balance, error)
	Withdraw(ctx context.Context, withdrawal *withdrawal.Withdrawal) error
}
