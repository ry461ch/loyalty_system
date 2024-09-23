package ordermemstorage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
)

func TestInsertOrder(t *testing.T) {
	accrual := float64(500)
	existingUserID := uuid.New()
	existingOrder := order.Order{
		ID:        "1115",
		Status:    order.PROCESSED,
		CreatedAt: time.Now().UTC(),
		Accrual:   &accrual,
	}

	testCases := []struct {
		testName          string
		userID            uuid.UUID
		orderID           string
		expectedOrdersNum int
	}{
		{
			testName:          "existing user",
			userID:            existingUserID,
			orderID:           "1321",
			expectedOrdersNum: 2,
		},
		{
			testName:          "new user",
			userID:            uuid.New(),
			orderID:           "1321",
			expectedOrdersNum: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewOrderMemStorage()
			storage.ordersToUsersMap.Store(existingOrder.ID, existingUserID)
			storage.usersToOrdersMap.Store(
				existingUserID,
				map[string]order.Order{
					existingOrder.ID: existingOrder,
				},
			)

			storage.InsertOrder(context.TODO(), tc.userID, tc.orderID, nil)
			val, _ := storage.usersToOrdersMap.Load(tc.userID)
			userOrders := val.(map[string]order.Order)
			assert.Equal(t, tc.expectedOrdersNum, len(userOrders), "num of orders doesn't match")
		})
	}
}

func TestGetUserID(t *testing.T) {
	accrual := float64(500)
	existingUserID := uuid.New()
	existingOrder := order.Order{
		ID:        "1115",
		Status:    order.PROCESSED,
		CreatedAt: time.Now().UTC(),
		Accrual:   &accrual,
	}

	testCases := []struct {
		testName       string
		orderID        string
		expectedUserID *uuid.UUID
	}{
		{
			testName:       "existing order",
			orderID:        "1115",
			expectedUserID: &existingUserID,
		},
		{
			testName:       "new order",
			orderID:        "1321",
			expectedUserID: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewOrderMemStorage()
			storage.ordersToUsersMap.Store(existingOrder.ID, existingUserID)
			storage.usersToOrdersMap.Store(
				existingUserID,
				map[string]order.Order{
					existingOrder.ID: existingOrder,
				},
			)

			userID, err := storage.GetOrderUserID(context.TODO(), tc.orderID)
			if tc.expectedUserID != nil {
				assert.Equal(t, *tc.expectedUserID, *userID, "users don't match")
			} else {
				assert.ErrorIs(t, err, exceptions.NewOrderNotFoundError(), "order found but shoildn't")
			}
		})
	}
}

func TestGetUserOrders(t *testing.T) {
	accrual := float64(500)
	existingUserID := uuid.New()
	existingOrders := []order.Order{
		{
			ID:        "1115",
			Status:    order.PROCESSED,
			CreatedAt: time.Now().UTC(),
			Accrual:   &accrual,
		},
		{
			ID:        "1321",
			Status:    order.INVALID,
			CreatedAt: time.Now().UTC(),
		},
	}

	testCases := []struct {
		testName          string
		userID            uuid.UUID
		expectedOrdersNum int
	}{
		{
			testName:          "existing user",
			userID:            existingUserID,
			expectedOrdersNum: 2,
		},
		{
			testName:          "new user",
			userID:            uuid.New(),
			expectedOrdersNum: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewOrderMemStorage()
			existingOrdersMap := map[string]order.Order{}
			for _, existingOrder := range existingOrders {
				storage.ordersToUsersMap.Store(existingOrder.ID, existingUserID)
				existingOrdersMap[existingOrder.ID] = existingOrder
			}
			storage.usersToOrdersMap.Store(existingUserID, existingOrdersMap)

			userOrders, _ := storage.GetUserOrders(context.TODO(), tc.userID)
			assert.Equal(t, tc.expectedOrdersNum, len(userOrders), "num of orders not equal")
		})
	}
}

func TestGetWaitingOrders(t *testing.T) {
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

	storage := NewOrderMemStorage()
	for existingOrderID := range existingOrdersUser1 {
		storage.ordersToUsersMap.Store(existingOrderID, existingUser1ID)
	}
	for existingOrderID := range existingOrdersUser2 {
		storage.ordersToUsersMap.Store(existingOrderID, existingUser2ID)
	}
	storage.usersToOrdersMap.Store(existingUser1ID, existingOrdersUser1)
	storage.usersToOrdersMap.Store(existingUser2ID, existingOrdersUser2)

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
			waitingOrders, _ := storage.GetWaitingOrders(context.TODO(), tc.limit, tc.createdAt)
			assert.Equal(t, len(tc.expectedOrderIDs), len(waitingOrders), "num of orders don't match")
			for _, userOrder := range waitingOrders {
				assert.Contains(t, tc.expectedOrderIDs, userOrder.ID, "user orders doesn't contain expected order id")
			}
		})
	}
}

func TestUpdateOrder(t *testing.T) {
	accrual := float64(500)
	createdAt, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	existingUserID := uuid.New()
	existingOrder := order.Order{
		ID:        "1115",
		Status:    order.PROCESSING,
		CreatedAt: createdAt,
	}

	testCases := []struct {
		testName    string
		newOrder    order.Order
		expectedErr error
	}{
		{
			testName: "existing order",
			newOrder: order.Order{
				ID:        existingOrder.ID,
				Status:    order.PROCESSED,
				Accrual:   &accrual,
				CreatedAt: createdAt,
			},
			expectedErr: nil,
		},
		{
			testName: "not existing order",
			newOrder: order.Order{
				ID:        "1321",
				Status:    order.PROCESSED,
				Accrual:   &accrual,
				CreatedAt: createdAt,
			},
			expectedErr: exceptions.NewOrderNotFoundError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewOrderMemStorage()
			storage.ordersToUsersMap.Store(existingOrder.ID, existingUserID)
			storage.usersToOrdersMap.Store(
				existingUserID,
				map[string]order.Order{
					existingOrder.ID: existingOrder,
				},
			)

			userID, err := storage.UpdateOrder(context.TODO(), &tc.newOrder, nil)
			assert.ErrorIs(t, tc.expectedErr, err, "errors don't match")
			val, _ := storage.usersToOrdersMap.Load(existingUserID)
			userOrders := val.(map[string]order.Order)
			if tc.expectedErr != nil {
				assert.Equal(t, existingOrder, userOrders[existingOrder.ID], "order was updated, but shouldn't")
			} else {
				assert.Equal(t, existingUserID, *userID, "user_ids not equal")
				assert.Equal(t, tc.newOrder, userOrders[existingOrder.ID], "order was updated, but shouldn't")
			}
		})
	}
}
