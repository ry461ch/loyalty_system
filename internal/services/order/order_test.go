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

			ordersInDB, _ := orderStorage.GetUserOrders(context.TODO(), tc.userID)
			assert.Equal(t, tc.expectedOrdersNum, len(ordersInDB), "orders num don't match")
		})
	}
}

func TestGetUserOrders(t *testing.T) {
	accrual := float64(500)
	existingUserID := uuid.New()
	existingOrders := []order.Order{
		{
			Accrual:   &accrual,
			Status:    order.PROCESSED,
			ID:        "1115",
			CreatedAt: time.Now().UTC(),
		},
		{
			Status:    order.NEW,
			ID:        "1321",
			CreatedAt: time.Now().UTC(),
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

			userOrdersList, _ := orderService.GetUserOrders(context.TODO(), tc.userID)
			userOrders := map[string]order.Order{}
			for _, updatedOrder := range userOrdersList {
				userOrders[updatedOrder.ID] = updatedOrder
			}

			for _, expectedOrder := range tc.expectedOrders {
				updatedOrder, ok := userOrders[expectedOrder.ID]
				assert.True(t, ok, "orser was not in updated list")
				assert.Equal(t, updatedOrder.Status, expectedOrder.Status, "statuses not equal")
				if expectedOrder.Accrual != nil {
					assert.Equal(t, expectedOrder.Accrual, updatedOrder.Accrual, "accrual nor equal")
				}
			}
		})
	}
}

func TestGetWaitingOrderIDs(t *testing.T) {
	accrual := float64(500)
	existingUser1ID := uuid.New()
	createdAt1, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	createdAt2, _ := time.Parse(time.RFC3339, "2020-12-10T16:09:53Z")
	createdAt3, _ := time.Parse(time.RFC3339, "2020-12-11T16:09:53Z")
	createdAt4, _ := time.Parse(time.RFC3339, "2020-12-12T16:09:53Z")
	existingOrdersUser1 := map[string]order.Order{
		"1115": {
			ID:        "1115",
			Status:    order.PROCESSING,
			CreatedAt: createdAt1,
		},
		"1321": {
			ID:        "1321",
			Status:    order.INVALID,
			CreatedAt: createdAt2,
		},
	}
	existingUser2ID := uuid.New()
	existingOrdersUser2 := map[string]order.Order{
		"1124": {
			ID:        "1124",
			Status:    order.PROCESSED,
			CreatedAt: createdAt3,
			Accrual:   &accrual,
		},
		"1131": {
			ID:        "1131",
			Status:    order.NEW,
			CreatedAt: createdAt4,
		},
	}

	orderStorage := ordermemstorage.NewOrderMemStorage()
	for _, existingOrder := range existingOrdersUser1 {
		orderStorage.InsertOrder(context.TODO(), existingUser1ID, existingOrder.ID, nil)
		orderStorage.UpdateOrder(context.TODO(), &existingOrder, nil)
	}
	for _, existingOrder := range existingOrdersUser2 {
		orderStorage.InsertOrder(context.TODO(), existingUser2ID, existingOrder.ID, nil)
		orderStorage.UpdateOrder(context.TODO(), &existingOrder, nil)
	}

	balanceStorage := balancememstorage.NewBalanceMemStorage()
	withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	orderService := NewOrderService(orderStorage, moneyService)

	requestCreatedAt1, _ := time.Parse(time.RFC3339, "2020-12-09T16:00:00Z")
	requestCreatedAt2, _ := time.Parse(time.RFC3339, "2020-12-10T16:00:00Z")
	requestCreatedAt3, _ := time.Parse(time.RFC3339, "2020-12-13T16:00:00Z")

	testCases := []struct {
		testName         string
		limit            int
		createdAt        *time.Time
		expectedOrderIDs []string
	}{
		{
			testName:         "all waiting orders",
			limit:            3,
			createdAt:        &requestCreatedAt3,
			expectedOrderIDs: []string{"1131", "1115"},
		},
		{
			testName:         "only one latest order",
			limit:            1,
			createdAt:        &requestCreatedAt3,
			expectedOrderIDs: []string{"1131"},
		},
		{
			testName:         "only one earliset order",
			limit:            2,
			createdAt:        &requestCreatedAt2,
			expectedOrderIDs: []string{"1115"},
		},
		{
			testName:         "too early createdAt",
			limit:            2,
			createdAt:        &requestCreatedAt1,
			expectedOrderIDs: []string{},
		},
		{
			testName:         "nil createdAt",
			limit:            1,
			createdAt:        nil,
			expectedOrderIDs: []string{"1131"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			waitingOrders, _ := orderService.GetWaitingOrders(context.TODO(), tc.limit, tc.createdAt)
			assert.Equal(t, len(tc.expectedOrderIDs), len(waitingOrders), "num of orders don't match")
			for _, userOrder := range waitingOrders {
				assert.Contains(t, tc.expectedOrderIDs, userOrder.ID, "user orders doesn't contain expected order id")
			}
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
				ordersInDB, _ := orderStorage.GetUserOrders(context.TODO(), existingUserID)
				assert.Equal(t, tc.inputOrder, ordersInDB[0], "orders not equal")
			} else {
				assert.ErrorIs(t, err, tc.expectedSavingResult, "exceptions don't match")
			}

			balanceInDB, _ := balanceStorage.GetBalance(context.TODO(), existingUserID)
			assert.Equal(t, tc.expectedBalance, *balanceInDB, "balances not equal")
		})
	}
}
