package ordermemstorage

import (
	"context"
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

func (oms *OrderMemStorage) InitializeOrderMemStorage(ctx context.Context) error {
	return nil
}

func (oms *OrderMemStorage) GetOrderUserId(ctx context.Context, orderId string) (*uuid.UUID, error) {
	val, ok := oms.ordersToUsersMap.Load(orderId)
	if !ok {
		return nil, exceptions.NewOrderNotFoundError()
	}
	userId := val.(uuid.UUID)
	return &userId, nil
}

func (oms *OrderMemStorage) InsertOrder(ctx context.Context, userId uuid.UUID, orderId string, trx *transaction.Trx) error {
	oms.ordersToUsersMap.Store(orderId, userId)
	newOrder := order.Order{
		Id:        orderId,
		Status:    order.NEW,
		CreatedAt: time.Now(),
	}

	val, ok := oms.usersToOrdersMap.Load(userId)
	if !ok {
		val = map[string]order.Order{}
	}
	userOrders := val.(map[string]order.Order)
	userOrders[orderId] = newOrder
	oms.usersToOrdersMap.Store(userId, userOrders)
	return nil
}

func (oms *OrderMemStorage) UpdateOrder(ctx context.Context, newOrder *order.Order, trx *transaction.Trx) error {
	val, ok := oms.ordersToUsersMap.Load(newOrder.Id)
	if !ok {
		return exceptions.NewOrderNotFoundError()
	}
	userId := val.(uuid.UUID)

	val, ok = oms.usersToOrdersMap.Load(userId)
	if !ok {
		return exceptions.NewOrderNotFoundError()
	}
	userOrders := val.(map[string]order.Order)
	userOrders[newOrder.Id] = *newOrder
	oms.usersToOrdersMap.Store(userId, userOrders)
	return nil
}

func (oms *OrderMemStorage) GetOrders(ctx context.Context, userId uuid.UUID) ([]order.Order, error) {
	val, ok := oms.usersToOrdersMap.Load(userId)
	if !ok {
		return []order.Order{}, nil
	}

	ordersList := []order.Order{}
	userOrders := val.(map[string]order.Order)
	for _, userOrder := range userOrders {
		ordersList = append(ordersList, userOrder)
	}
	return ordersList, nil
}

func (*OrderMemStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, nil)
}
