package components

import (
	"context"

	"github.com/ry461ch/loyalty_system/internal/models/order"
)

type OrderUpdater interface {
	UpdateOrders(ctx context.Context, updatedOrders <-chan order.Order)
}

type OrderGetter interface {
	GetWaitingOrderIDs(ctx context.Context, orderIDsChannel chan<- string)
}

type OrderSender interface {
	GetUpdatedOrders(ctx context.Context, orderIDsChannel <-chan string, updatedOrders chan<- order.Order)
}
