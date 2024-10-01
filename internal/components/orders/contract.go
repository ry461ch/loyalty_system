package ordercomponents

import (
	"github.com/ry461ch/loyalty_system/internal/components/orders/getter"
	"github.com/ry461ch/loyalty_system/internal/components/orders/updater"
)

type OrderService interface {
	orderupdater.OrderUpdaterService
	ordergetter.WaitingOrdersGetterService
}
