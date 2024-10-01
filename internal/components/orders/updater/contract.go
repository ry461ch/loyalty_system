package orderupdater

import (
	"context"

	"github.com/ry461ch/loyalty_system/internal/models/order"
)

type OrderUpdaterService interface {
	UpdateOrder(ctx context.Context, inputOrder *order.Order) error
}
