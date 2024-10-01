package orderservice

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/helpers/transaction"
	"github.com/ry461ch/loyalty_system/internal/models/order"
)

type OrderStorage interface {
	GetOrderUserID(ctx context.Context, orderID string) (*uuid.UUID, error)
	InsertOrder(ctx context.Context, userID uuid.UUID, orderID string, trx *transaction.Trx) error
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error)
	GetWaitingOrders(ctx context.Context, limit int, inputCreatedAt *time.Time) ([]order.Order, error)
	UpdateOrder(ctx context.Context, order *order.Order, tx *transaction.Trx) (*uuid.UUID, error)
	BeginTx(ctx context.Context) (*transaction.Trx, error)
}

type AccrualAdderService interface {
	AddAccrual(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error
}
