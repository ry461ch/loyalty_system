package handlersimpl

import (
	"github.com/ry461ch/loyalty_system/internal/services"
	"github.com/ry461ch/loyalty_system/internal/handlers/money"
	"github.com/ry461ch/loyalty_system/internal/handlers/orders"
	"github.com/ry461ch/loyalty_system/internal/handlers/auth"
)

type Handlers struct {
	AuthHandlers *authhandlers.AuthHandlers
	MoneyHandlers *moneyhandlers.MoneyHandlers
	OrdersHandlers *orderhandlers.OrderHandlers
}

func NewHandlers(
	moneyService services.MoneyService,
	orderService services.OrderService,
	userService services.UserService,
) *Handlers {
	return &Handlers{
		AuthHandlers: authhandlers.NewAuthHandlers(userService),
		MoneyHandlers: moneyhandlers.NewMoneyHandlers(moneyService),
		OrdersHandlers: orderhandlers.NewOrderHandlers(orderService),
	}
}
