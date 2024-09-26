package handlers

import (
	"github.com/ry461ch/loyalty_system/internal/handlers/auth"
	"github.com/ry461ch/loyalty_system/internal/handlers/money"
	"github.com/ry461ch/loyalty_system/internal/handlers/orders"
)

type Handlers struct {
	AuthHandlers   *authhandlers.AuthHandlers
	MoneyHandlers  *moneyhandlers.MoneyHandlers
	OrdersHandlers *orderhandlers.OrderHandlers
}

func NewHandlers(
	moneyService moneyhandlers.MoneyService,
	orderService orderhandlers.OrderService,
	userService authhandlers.UserService,
) *Handlers {
	return &Handlers{
		AuthHandlers:   authhandlers.NewAuthHandlers(userService),
		MoneyHandlers:  moneyhandlers.NewMoneyHandlers(moneyService),
		OrdersHandlers: orderhandlers.NewOrderHandlers(orderService),
	}
}
