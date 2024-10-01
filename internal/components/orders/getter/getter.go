package ordergetter

import (
	"context"
	"errors"
	"time"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderGetter struct {
	orderService   WaitingOrdersGetterService
	getOrdersLimit int
	rateLimit      int
}

func NewOrderGetter(orderService WaitingOrdersGetterService, cfg *config.Config) *OrderGetter {
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

func (og *OrderGetter) getWaitingOrderIDs(ctx context.Context, orderIDsChannel chan<- string) {
	logging.Logger.Infof("Order Getter: initiated")
	var createdAt *time.Time

	for {
		select {
		case <-ctx.Done():
			logging.Logger.Infof("Order Getter: gracefully shutdown")
			return
		default:
		}

		var err error
		createdAt, err = og.getWaitingOrderIDsIteration(ctx, orderIDsChannel, createdAt)
		if err != nil {
			if errors.Is(err, exceptions.ErrGracefullyShutDown) {
				logging.Logger.Infof("Order Getter: gracefully shutdown")
				return
			}
			logging.Logger.Errorf("Order Getter: exceptions occured while getting waiting orders: %s", err.Error())
			return
		}

		if createdAt == nil {
			logging.Logger.Infof("Order Getter: gracefully shutdown")
			return
		}

		time.Sleep(time.Second / time.Duration(og.rateLimit))
	}
}

func (og *OrderGetter) GetWaitingOrderIdsGenerator(ctx context.Context) chan string {
	orderIDsChannel := make(chan string)

	go func() {
		defer close(orderIDsChannel)
		og.getWaitingOrderIDs(ctx, orderIDsChannel)
	}()

	return orderIDsChannel
}
