package services

import (
	"github.com/ry461ch/loyalty_system/internal/services/money"
	"github.com/ry461ch/loyalty_system/internal/services/order"
	"github.com/ry461ch/loyalty_system/internal/services/user"
	"github.com/ry461ch/loyalty_system/pkg/authentication"
)

type Services struct {
	UserService  *userservice.UserService
	MoneyService *moneyservice.MoneyService
	OrderService *orderservice.OrderService
}

func NewServices(
	balanceStorage moneyservice.BalanceStorage,
	withdrawalStorage moneyservice.WithdrawalStorage,
	userStorage userservice.UserStorage,
	orderStorage orderservice.OrderStorage,
	authenticator *authentication.Authenticator,
) *Services {
	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	return &Services{
		UserService:  userservice.NewUserService(userStorage, authenticator),
		MoneyService: moneyService,
		OrderService: orderservice.NewOrderService(orderStorage, moneyService),
	}
}
