package orderupdater

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/services/money"
	"github.com/ry461ch/loyalty_system/internal/services/order"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

func TestUpdater(t *testing.T) {
	existingUserID := uuid.New()
	accrual := float64(200)
	expectedOrders := []order.Order{
		{
			ID: "1115",
			Status: order.NEW,
		},
		{
			ID: "1321",
			Status: order.INVALID,
		},
		{
			ID: "1214",
			Status: order.PROCESSED,
			Accrual: &accrual,
		},
	}
	existingBalance := balance.Balance{
		Current:   200,
		Withdrawn: 200,
	}

	orderStorage := ordermemstorage.NewOrderMemStorage()
	balanceStorage := balancememstorage.NewBalanceMemStorage()
	withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
	for _, existingOrder := range expectedOrders {
		orderStorage.InsertOrder(context.TODO(), existingUserID, existingOrder.ID, nil)
	}
	balanceStorage.AddBalance(context.TODO(), existingUserID, existingBalance.Current+existingBalance.Withdrawn, nil)
	balanceStorage.ReduceBalance(context.TODO(), existingUserID, existingBalance.Withdrawn, nil)

	updatedOrdersChannel := make(chan order.Order, 3)
	for _, expectedOrder := range expectedOrders {
		updatedOrdersChannel <- expectedOrder
	}
	close(updatedOrdersChannel)

	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	orderService := orderservice.NewOrderService(orderStorage, moneyService)
	updater := OrderUpdater{
		orderService: orderService,
		workersNum: 2,
	}

	start := time.Now()
	updater.UpdateOrders(context.TODO(), updatedOrdersChannel)
	assert.GreaterOrEqual(t, time.Since(start), time.Second * 2, "workers worked less than 2 seconds")

	updatedOrdersList, _ := orderStorage.GetUserOrders(context.TODO(), existingUserID)

	updatedOrders := map[string]order.Order{}
	for _, updatedOrder := range updatedOrdersList {
		updatedOrders[updatedOrder.ID] = updatedOrder
	}

	for _, expectedOrder := range expectedOrders {
		updatedOrder, ok := updatedOrders[expectedOrder.ID]
		assert.True(t, ok, "orser was not in updated list")
		assert.Equal(t, updatedOrder.Status, expectedOrder.Status, "statuses not equal")
		if expectedOrder.Accrual != nil {
			assert.Equal(t, expectedOrder.Accrual, updatedOrder.Accrual, "accrual nor equal")
		}
	}

	expectedBalance := balance.Balance{
		Current: existingBalance.Current + accrual,
		Withdrawn: existingBalance.Withdrawn,
	}
	userBalance, _ := balanceStorage.GetBalance(context.TODO(), existingUserID)
	assert.Equal(t, expectedBalance, *userBalance, "balances not equal")
}
