package ordergetter

import (
	"context"
	"time"

	"github.com/ry461ch/loyalty_system/internal/models/order"
)

type WaitingOrdersGetterService interface {
	GetWaitingOrders(ctx context.Context, limit int, createdAt *time.Time) ([]order.Order, error)
}
