package ordercomponents

import (
	"github.com/ry461ch/loyalty_system/internal/components/orders/sender"
	"github.com/ry461ch/loyalty_system/internal/components/orders/updater"
	"github.com/ry461ch/loyalty_system/internal/components/orders/getter"
	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/services"
)

type OrderComponents struct {
	Sender *ordersender.OrderSender
	Getter *ordergetter.OrderGetter
	Updater *orderupdater.OrderUpdater
}

func NewOrderComponents(cfg *config.Config, orderService services.OrderService) *OrderComponents {
	return &OrderComponents{
		Sender: ordersender.NewOrderSender(cfg),
		Getter: ordergetter.NewOrderGetter(orderService, cfg),
		Updater: orderupdater.NewOrderUpdater(orderService, cfg),
	}
}
