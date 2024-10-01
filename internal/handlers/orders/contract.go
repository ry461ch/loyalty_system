package orderhandlers

import (
	"context"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/order"
)

type OrderService interface {
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error)
	InsertOrder(ctx context.Context, userID uuid.UUID, orderID string) error
}
