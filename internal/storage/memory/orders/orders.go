package ordermemstorage

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type OrderMemStorage struct {
	usersToOrdersMap sync.Map // map[uuid.UUID]map[string]order.Order
	ordersToUsersMap sync.Map // map[string]uuid.UUID
}

func NewOrderMemStorage() *OrderMemStorage {
	return &OrderMemStorage{}
}

func (oms *OrderMemStorage) GetOrderUserID(ctx context.Context, orderID string) (*uuid.UUID, error) {
	val, ok := oms.ordersToUsersMap.Load(orderID)
	if !ok {
		return nil, exceptions.NewOrderNotFoundError()
	}
	userID := val.(uuid.UUID)
	return &userID, nil
}

func (oms *OrderMemStorage) InsertOrder(ctx context.Context, userID uuid.UUID, orderID string, trx *transaction.Trx) error {
	oms.ordersToUsersMap.Store(orderID, userID)
	newOrder := order.Order{
		ID:        orderID,
		Status:    order.NEW,
		CreatedAt: time.Now().UTC(),
	}

	val, ok := oms.usersToOrdersMap.Load(userID)
	if !ok {
		val = map[string]order.Order{}
	}
	userOrders := val.(map[string]order.Order)
	userOrders[orderID] = newOrder
	oms.usersToOrdersMap.Store(userID, userOrders)
	return nil
}

func (oms *OrderMemStorage) UpdateOrder(ctx context.Context, newOrder *order.Order, trx *transaction.Trx) (*uuid.UUID, error) {
	val, ok := oms.ordersToUsersMap.Load(newOrder.ID)
	if !ok {
		return nil, exceptions.NewOrderNotFoundError()
	}
	userID := val.(uuid.UUID)

	val, ok = oms.usersToOrdersMap.Load(userID)
	if !ok {
		return nil, exceptions.NewOrderNotFoundError()
	}
	userOrders := val.(map[string]order.Order)
	userOrders[newOrder.ID] = *newOrder
	oms.usersToOrdersMap.Store(userID, userOrders)
	return &userID, nil
}

func (oms *OrderMemStorage) GetWaitingOrders(ctx context.Context, limit int, inputCreatedAt *time.Time) ([]order.Order, error) {
	var waitingOrders []order.Order
	oms.usersToOrdersMap.Range(func(key any, val any) bool {
		userOrders := val.(map[string]order.Order)
		for _, userOrder := range userOrders {
			if (userOrder.Status == order.NEW || userOrder.Status == order.PROCESSING) &&
				(inputCreatedAt == nil || inputCreatedAt.Compare(userOrder.CreatedAt) > 0) {
				waitingOrders = append(waitingOrders, userOrder)
			}
		}

		return true
	})

	slices.SortFunc(waitingOrders, func(left, right order.Order) int {
		return right.CreatedAt.Compare(left.CreatedAt)
	})

	resultOrders := waitingOrders[:min(limit, len(waitingOrders))]
	return resultOrders, nil
}

func (oms *OrderMemStorage) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error) {
	val, ok := oms.usersToOrdersMap.Load(userID)
	if !ok {
		return []order.Order{}, nil
	}

	ordersList := []order.Order{}
	userOrders := val.(map[string]order.Order)
	for _, userOrder := range userOrders {
		ordersList = append(ordersList, userOrder)
	}
	slices.SortFunc(ordersList, func(left, right order.Order) int {
		return right.CreatedAt.Compare(left.CreatedAt)
	})
	return ordersList, nil
}

func (*OrderMemStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, nil)
}
