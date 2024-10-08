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
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

func TestGetter(t *testing.T) {
	logging.Initialize("INFO")
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

	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	orderService := orderservice.NewOrderService(orderStorage, moneyService)
	getter := OrderGetter{
		orderService:   orderService,
		getOrdersLimit: 2,
		rateLimit:      1,
	}

	start := time.Now().UTC()
	orderIDsChannel := getter.GetWaitingOrderIdsGenerator(context.TODO())

	updatedOrders := map[string]bool{}
	for updatedOrderID := range orderIDsChannel {
		updatedOrders[updatedOrderID] = true
	}
	assert.GreaterOrEqual(t, time.Since(start), time.Second, "workers worked less than 1 second")

	for _, expectedOrder := range expectedOrders {
		if expectedOrder.Status == order.NEW || expectedOrder.Status == order.PROCESSING {
			_, ok := updatedOrders[expectedOrder.ID]
			assert.True(t, ok, "orser was not in updated list")
		}
	}
}
