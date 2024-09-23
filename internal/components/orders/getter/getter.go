package ordergetter

import (
	"context"
	"errors"
	"time"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/services"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderGetter struct {
	orderService   services.OrderService
	getOrdersLimit int
	rateLimit      int
}

func NewOrderGetter(orderService services.OrderService, cfg *config.Config) *OrderGetter {
	return &OrderGetter{
		orderService:   orderService,
		getOrdersLimit: cfg.OrderGetterOrdersLimit,
		rateLimit:      cfg.OrderGetterRateLimit,
	}
}

func (og *OrderGetter) getWaitingOrderIDsIteration(ctx context.Context, orderIDsChannel chan<- string, createdAt *time.Time) (*time.Time, error) {
	waitingOrders, err := og.orderService.GetWaitingOrders(ctx, og.getOrdersLimit, createdAt)
	if err != nil {
		logging.Logger.Warnf("OrderGetter: exceptions occured while getting waiting orders: %s", err.Error())
		return nil, err
	}

	for _, waitingOrder := range waitingOrders {
		select {
		case <-ctx.Done():
			return nil, errors.New("graceful shutdown getter waiting orders")
		case orderIDsChannel <- waitingOrder.ID:
		}
	}

	if len(waitingOrders) < og.getOrdersLimit {
		return nil, nil
	}

	return &waitingOrders[len(waitingOrders)-1].CreatedAt, nil
}

func (og *OrderGetter) GetWaitingOrderIDs(ctx context.Context, orderIDsChannel chan<- string) error {
	var createdAt *time.Time
	ticker := time.NewTicker(time.Second / time.Duration(og.rateLimit))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("graceful shutdown getter waiting orders")
		case <-ticker.C:
			var err error
			createdAt, err = og.getWaitingOrderIDsIteration(ctx, orderIDsChannel, createdAt)
			if err != nil {
				return err
			}
		}

		if createdAt == nil {
			return nil
		}
	}
}
