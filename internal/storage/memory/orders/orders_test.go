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

func TestGetOrders(t *testing.T) {
	accrual := float64(500)
	existingUserID := uuid.New()
	existingOrders := map[string]order.Order{
		"1115": {
			ID:        "1115",
			Status:    order.PROCESSED,
			CreatedAt: time.Now().UTC(),
			Accrual:   &accrual,
		},
		"1321": {
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
			for existingOrderID := range existingOrders {
				storage.ordersToUsersMap.Store(existingOrderID, existingUserID)
			}
			storage.usersToOrdersMap.Store(existingUserID, existingOrders)

			userOrders, _ := storage.GetOrders(context.TODO(), tc.userID)
			assert.Equal(t, tc.expectedOrdersNum, len(userOrders), "num of orders not equal")
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

			err := storage.UpdateOrder(context.TODO(), &tc.newOrder, nil)
			assert.ErrorIs(t, tc.expectedErr, err, "errors don't match")
			val, _ := storage.usersToOrdersMap.Load(existingUserID)
			userOrders := val.(map[string]order.Order)
			if tc.expectedErr != nil {
				assert.Equal(t, existingOrder, userOrders[existingOrder.ID], "order was updated, but shouldn't")
			} else {
				assert.Equal(t, tc.newOrder, userOrders[existingOrder.ID], "order was updated, but shouldn't")
			}
		})
	}
}
