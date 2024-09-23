package ordergetter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/services/money"
	"github.com/ry461ch/loyalty_system/internal/services/order"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

func TestGetter(t *testing.T) {
	expectedOrders := []order.Order{
		{
			ID:     "1115",
			Status: order.NEW,
		},
		{
			ID:     "1321",
			Status: order.INVALID,
		},
		{
			ID:     "1214",
			Status: order.PROCESSING,
		},
		{
			ID:     "1313",
			Status: order.NEW,
		},
	}

	orderStorage := ordermemstorage.NewOrderMemStorage()
	balanceStorage := balancememstorage.NewBalanceMemStorage()
	withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
	for _, existingOrder := range expectedOrders {
		orderStorage.InsertOrder(context.TODO(), uuid.New(), existingOrder.ID, nil)
		if existingOrder.Status != order.NEW {
			orderStorage.UpdateOrder(context.TODO(), &existingOrder, nil)
		}
	}

	orderIDsChannel := make(chan string, 3)
	defer close(orderIDsChannel)

	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	orderService := orderservice.NewOrderService(orderStorage, moneyService)
	getter := OrderGetter{
		orderService:   orderService,
		getOrdersLimit: 2,
		rateLimit:      1,
	}

	start := time.Now().UTC()
	getter.GetWaitingOrderIDs(context.TODO(), orderIDsChannel)
	assert.GreaterOrEqual(t, time.Since(start), time.Second*2, "workers worked less than 2 seconds")

	updatedOrders := map[string]bool{}
	for range 3 {
		updatedOrderID := <-orderIDsChannel
		updatedOrders[updatedOrderID] = true
	}

	for _, expectedOrder := range expectedOrders {
		if expectedOrder.Status == order.NEW || expectedOrder.Status == order.PROCESSING {
			_, ok := updatedOrders[expectedOrder.ID]
			assert.True(t, ok, "orser was not in updated list")
		}
	}
}
