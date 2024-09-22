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

func (os *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]order.Order, error) {
	return os.orderStorage.GetUserOrders(ctx, userID)
}

func (os *OrderService) GetWaitingOrderIDs(ctx context.Context, limit, offset int) ([]string, error) {
	return os.orderStorage.GetWaitingOrderIDs(ctx, limit, offset)
}

func (os *OrderService) InsertOrder(ctx context.Context, userID uuid.UUID, orderID string) error {
	if !orderhelpers.ValidateOrderID(orderID) {
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
	err = os.orderStorage.InsertOrder(ctx, userID, orderID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (os *OrderService) UpdateOrder(ctx context.Context, inputOrder *order.Order) error {
	if inputOrder == nil {
		return errors.New("invalid input order")
	}

	tx, err := os.orderStorage.BeginTx(ctx)
	if err != nil {
		return err
	}

	userID, err := os.orderStorage.UpdateOrder(ctx, inputOrder, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	if userID == nil {
		tx.Rollback()
		return errors.New("update order returns empty userID")
	}

	if inputOrder.Accrual == nil {
		tx.Commit()
		return nil
	}

	err = os.moneyService.AddAccrual(ctx, *userID, *inputOrder.Accrual, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	tx.Commit()
	return nil
}
