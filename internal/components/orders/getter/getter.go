package ordergetter

import (
	"context"
	"errors"
	"time"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/interfaces/services"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderGetter struct {
	orderService   services.WaitingOrdersGetterService
	getOrdersLimit int
	rateLimit      int
}

func NewOrderGetter(orderService services.WaitingOrdersGetterService, cfg *config.Config) *OrderGetter {
	return &OrderGetter{
		orderService:   orderService,
		getOrdersLimit: cfg.OrderGetterOrdersLimit,
		rateLimit:      cfg.OrderGetterRateLimit,
	}
}

func (og *OrderGetter) getWaitingOrderIDsIteration(ctx context.Context, orderIDsChannel chan<- string, createdAt *time.Time) (*time.Time, error) {
	waitingOrders, err := og.orderService.GetWaitingOrders(ctx, og.getOrdersLimit, createdAt)
	if createdAt != nil {
		logging.Logger.Infof("Order Getter: got %d orders with createdAt less than %s", len(waitingOrders), createdAt.String())
	} else {
		logging.Logger.Infof("Order Getter: got %d orders", len(waitingOrders))
	}

	if err != nil {
		return nil, err
	}

	for _, waitingOrder := range waitingOrders {
		select {
		case <-ctx.Done():
			return nil, exceptions.ErrGracefullyShutDown
		case orderIDsChannel <- waitingOrder.ID:
		}
	}

	if len(waitingOrders) < og.getOrdersLimit {
		return nil, nil
	}

	return &waitingOrders[len(waitingOrders)-1].CreatedAt, nil
}

func (og *OrderGetter) GetWaitingOrderIDs(ctx context.Context, orderIDsChannel chan<- string) error {
	logging.Logger.Infof("Order Getter: initiated")
	var createdAt *time.Time
	ticker := time.NewTicker(time.Second / time.Duration(og.rateLimit))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logging.Logger.Infof("Order Getter: gracefully shutdown")
			return nil
		case <-ticker.C:
			var err error
			createdAt, err = og.getWaitingOrderIDsIteration(ctx, orderIDsChannel, createdAt)
			if err != nil {
				if errors.Is(err, exceptions.ErrGracefullyShutDown) {
					logging.Logger.Infof("Order Getter: gracefully shutdown")
					return nil
				}
				logging.Logger.Errorf("Order Getter: exceptions occured while getting waiting orders: %s", err.Error())
				return err
			}
		}

		if createdAt == nil {
			logging.Logger.Infof("Order Getter: gracefully shutdown")
			return nil
		}
	}
}
