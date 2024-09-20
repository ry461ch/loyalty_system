package orderservice

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/services/money"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

func TestSaveOrder(t *testing.T) {
	existingUserID := uuid.New()
	existingOrderID := "1115"

	testCases := []struct {
		testName             string
		userID               uuid.UUID
		orderID              string
		expectedSavingResult error
		expectedOrdersNum    int
	}{
		{
			testName:             "successfully saved with same user",
			userID:               existingUserID,
			orderID:              "1321",
			expectedSavingResult: nil,
			expectedOrdersNum:    2,
		},
		{
			testName:             "successfully saved with another user",
			userID:               uuid.New(),
			orderID:              "1321",
			expectedSavingResult: nil,
			expectedOrdersNum:    1,
		},
		{
			testName:             "invalid order id",
			userID:               uuid.New(),
			orderID:              "1322a",
			expectedSavingResult: exceptions.NewOrderBadIDFormatError(),
			expectedOrdersNum:    0,
		},
		{
			testName:             "existing order id with another user",
			userID:               uuid.New(),
			orderID:              existingOrderID,
			expectedSavingResult: exceptions.NewOrderConflictAnotherUserError(),
			expectedOrdersNum:    0,
		},
		{
			testName:             "existing order id with same user",
			userID:               existingUserID,
			orderID:              existingOrderID,
			expectedSavingResult: exceptions.NewOrderConflictSameUserError(),
			expectedOrdersNum:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			orderStorage := ordermemstorage.NewOrderMemStorage()
			orderStorage.InsertOrder(context.TODO(), existingUserID, existingOrderID, nil)
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
			orderService := NewOrderService(orderStorage, moneyService)

			err := orderService.InsertOrder(context.TODO(), tc.userID, tc.orderID)
			assert.ErrorIs(t, tc.expectedSavingResult, err, "exceptions don't match")

			ordersInDb, _ := orderStorage.GetOrders(context.TODO(), tc.userID)
			assert.Equal(t, tc.expectedOrdersNum, len(ordersInDb), "orders num don't match")
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	existingUserID := uuid.New()
	existingOrderID := "1115"
	existingBalance := balance.Balance{
		Current:   200,
		Withdrawn: 200,
	}
	accrual := float64(200)

	testCases := []struct {
		testName             string
		inputOrder           order.Order
		expectedSavingResult error
		expectedBalance      balance.Balance
	}{
		{
			testName: "successfully updated with accrual",
			inputOrder: order.Order{
				ID:      existingOrderID,
				Status:  order.PROCESSED,
				Accrual: &accrual,
			},
			expectedSavingResult: nil,
			expectedBalance: balance.Balance{
				Current:   existingBalance.Current + accrual,
				Withdrawn: existingBalance.Withdrawn,
			},
		},
		{
			testName: "successfully updated without accrual",
			inputOrder: order.Order{
				ID:     existingOrderID,
				Status: order.INVALID,
			},
			expectedSavingResult: nil,
			expectedBalance:      existingBalance,
		},
		{
			testName: "order not found",
			inputOrder: order.Order{
				ID:      "1123",
				Status:  order.PROCESSED,
				Accrual: &accrual,
			},
			expectedSavingResult: exceptions.NewOrderNotFoundError(),
			expectedBalance:      existingBalance,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			orderStorage := ordermemstorage.NewOrderMemStorage()
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			orderStorage.InsertOrder(context.TODO(), existingUserID, existingOrderID, nil)
			balanceStorage.AddBalance(context.TODO(), existingUserID, existingBalance.Current+existingBalance.Withdrawn, nil)
			balanceStorage.ReduceBalance(context.TODO(), existingUserID, existingBalance.Withdrawn, nil)
			moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
			orderService := NewOrderService(orderStorage, moneyService)

			err := orderService.UpdateOrder(context.TODO(), &tc.inputOrder)
			if tc.expectedSavingResult == nil {
				assert.Nil(t, err, "not expected error")
				ordersInDb, _ := orderStorage.GetOrders(context.TODO(), existingUserID)
				assert.Equal(t, tc.inputOrder, ordersInDb[0], "orders not equal")
			} else {
				assert.ErrorIs(t, err, tc.expectedSavingResult, "exceptions don't match")
			}

			balanceInDb, _ := balanceStorage.GetBalance(context.TODO(), existingUserID)
			assert.Equal(t, tc.expectedBalance, *balanceInDb, "balances not equal")
		})
	}
}

func TestGetOrders(t *testing.T) {
	accrual := float64(500)
	existingUserID := uuid.New()
	existingOrders := []order.Order{
		{
			Accrual:   &accrual,
			Status:    order.PROCESSED,
			ID:        "1115",
			CreatedAt: time.Now(),
		},
		{
			Status:    order.NEW,
			ID:        "1321",
			CreatedAt: time.Now(),
		},
	}

	testCases := []struct {
		testName       string
		userID         uuid.UUID
		expectedOrders []order.Order
	}{
		{
			testName:       "new user",
			userID:         uuid.New(),
			expectedOrders: []order.Order{},
		},
		{
			testName:       "existing user",
			userID:         existingUserID,
			expectedOrders: existingOrders,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			orderStorage := ordermemstorage.NewOrderMemStorage()
			for _, newOrder := range existingOrders {
				orderStorage.InsertOrder(context.TODO(), existingUserID, newOrder.ID, nil)
				orderStorage.UpdateOrder(context.TODO(), &newOrder, nil)
			}
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
			orderService := NewOrderService(orderStorage, moneyService)

			userOrders, _ := orderService.GetOrders(context.TODO(), tc.userID)
			assert.Equal(t, tc.expectedOrders, userOrders, "orders are not equal")
		})
	}
}
