package orderservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/helpers/order"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/services"
	"github.com/ry461ch/loyalty_system/internal/storage"
)

type OrderService struct {
	orderStorage storage.OrderStorage
	moneyService services.MoneyService
}

func NewOrderService(orderStorage storage.OrderStorage, moneyService services.MoneyService) *OrderService {
	return &OrderService{
		orderStorage: orderStorage,
		moneyService: moneyService,
	}
}

func (os *OrderService) GetOrders(ctx context.Context, userId uuid.UUID) ([]order.Order, error) {
	return os.orderStorage.GetOrders(ctx, userId)
}

func (os *OrderService) InsertOrder(ctx context.Context, userId uuid.UUID, orderId string) error {
	if !orderhelper.ValidateOrderId(orderId) {
		return exceptions.NewOrderBadIdFormatError()
	}

	orderUserId, err := os.orderStorage.GetOrderUserId(ctx, orderId)
	if orderUserId != nil {
		if orderUserId.String() == userId.String() {
			return exceptions.NewOrderConflictSameUserError()
		}
		return exceptions.NewOrderConflictAnotherUserError()
	}
	if err != nil && !errors.Is(err, exceptions.NewOrderNotFoundError()) {
		return err
	}

	tx, err := os.orderStorage.BeginTx(ctx)
	if err != nil {
		return err
	}
	return os.orderStorage.InsertOrder(ctx, userId, orderId, tx)
}

func (os *OrderService) UpdateOrder(ctx context.Context, updatedOrder *order.Order) error {
	orderUserId, err := os.orderStorage.GetOrderUserId(ctx, updatedOrder.Id)
	if err != nil {
		return err
	}

	tx, err := os.orderStorage.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = os.orderStorage.UpdateOrder(ctx, updatedOrder, tx)
	if err != nil {
		return err
	}

	if updatedOrder.Accrual == nil || updatedOrder.Status != order.PROCESSED {
		return nil
	}

	return os.moneyService.AddAccrual(ctx, *orderUserId, *updatedOrder.Accrual, tx)
}
