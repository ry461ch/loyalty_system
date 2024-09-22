package ordergetter

import (
	"context"
	"errors"
	"time"

	"github.com/ry461ch/loyalty_system/internal/services"
	"github.com/ry461ch/loyalty_system/pkg/logging"
	"github.com/ry461ch/loyalty_system/internal/config"
)

type OrderGetter struct {
	orderService services.OrderService
	getOrdersLimit int
	rateLimit int
}

func NewOrderGetter(orderService services.OrderService, cfg *config.Config) *OrderGetter {
	return &OrderGetter{
		orderService: orderService,
		getOrdersLimit: cfg.OrderGetterOrdersLimit,
		rateLimit: cfg.OrderGetterRateLimit,
	}
}

func (og *OrderGetter) getWaitingOrderIDsIteration(ctx context.Context, orderIDsChannel chan<- string, offset int) (bool, error) {
	waitingOrderIDs, err := og.orderService.GetWaitingOrderIDs(ctx, og.getOrdersLimit, offset)
	if err != nil {
		logging.Logger.Warnf("OrderUpdater: exceptions occured while getting waiting orders: %s", err.Error())
		return false, err
	}

	for _, orderID := range waitingOrderIDs {
		select {
		case <-ctx.Done():
			return false, errors.New("graceful shutdown getter waiting orders")
		case orderIDsChannel <- orderID:
		}
	}

	if len(waitingOrderIDs) < og.getOrdersLimit {
		return false, nil
	}

	return true, nil
}

func (og *OrderGetter) GetWaitingOrderIDs(ctx context.Context, orderIDsChannel chan<- string) error {
	shouldContinue := true
	offset := 0
	ticker := time.NewTicker(time.Second/time.Duration(og.rateLimit))
	defer ticker.Stop()

	for {
		select{
		case <-ctx.Done():
			return errors.New("graceful shutdown getter waiting orders")
		case <-ticker.C:
			var err error
			shouldContinue, err = og.getWaitingOrderIDsIteration(ctx, orderIDsChannel, offset)
			if err != nil {
				return err
			}

			offset += og.getOrdersLimit
		}

		if !shouldContinue {
			return nil
		}
	}
}
