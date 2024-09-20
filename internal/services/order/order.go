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

func (os *OrderService) GetOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error) {
	return os.orderStorage.GetOrders(ctx, userID)
}

func (os *OrderService) InsertOrder(ctx context.Context, userID uuid.UUID, orderID string) error {
	if !orderhelper.ValidateOrderID(orderID) {
		return exceptions.NewOrderBadIDFormatError()
	}

	orderUserID, err := os.orderStorage.GetOrderUserID(ctx, orderID)
	if orderUserID != nil {
		if orderUserID.String() == userID.String() {
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
	return os.orderStorage.InsertOrder(ctx, userID, orderID, tx)
}

func (os *OrderService) UpdateOrder(ctx context.Context, updatedOrder *order.Order) error {
	orderUserID, err := os.orderStorage.GetOrderUserID(ctx, updatedOrder.ID)
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

	return os.moneyService.AddAccrual(ctx, *orderUserID, *updatedOrder.Accrual, tx)
}
