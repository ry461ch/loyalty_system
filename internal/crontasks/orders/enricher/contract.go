package orderenricher

import (
	"context"

	"github.com/ry461ch/loyalty_system/internal/models/order"
)

type OrderUpdater interface {
	UpdateOrders(ctx context.Context, updatedOrders <-chan order.Order)
}

type OrderGetter interface {
	GetWaitingOrderIdsGenerator(ctx context.Context) chan string
}

type OrderSender interface {
	SendOrdersGenerator(ctx context.Context, orderIDsChannel <-chan string) chan order.Order
}
