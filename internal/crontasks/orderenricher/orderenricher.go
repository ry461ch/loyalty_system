package orderenricher

import (
	"context"
	"sync"
	"time"
	"errors"

	"github.com/ry461ch/loyalty_system/internal/components/sender"
	"github.com/ry461ch/loyalty_system/internal/components/updater"
	"github.com/ry461ch/loyalty_system/internal/components/getter"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderEnricher struct {
	orderSender ordersender.OrderSender
	orderUpdater orderupdater.OrderUpdater
	orderGetter ordergetter.OrderGetter
	iterationOrdersLimit int
	iterationTimeout time.Duration
	iterationPeriod  time.Duration
}

func NewOrderEnricher(
	orderGetter ordergetter.OrderGetter,
	orderSender ordersender.OrderSender,
	orderUpdater orderupdater.OrderUpdater,
) *OrderEnricher {
	return &OrderEnricher{
		orderGetter: orderGetter,
		orderSender: orderSender,
		orderUpdater: orderUpdater,
		iterationOrdersLimit: 1000,
		iterationTimeout: time.Minute,
		iterationPeriod: time.Minute,
	}
}
 
func (oe *OrderEnricher) runIteration(ctx context.Context) {
	orderIDsChannel := make(chan string, oe.iterationOrdersLimit)
	updatedOrders := make(chan order.Order, oe.iterationOrdersLimit)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		err := oe.orderGetter.GetWaitingOrderIDs(ctx, orderIDsChannel)
		if err != nil {
			logging.Logger.Error("error occured in order getter")
		}
		close(orderIDsChannel)
		wg.Done()
	}()

	go func() {
		err := oe.orderSender.GetUpdatedOrders(ctx, orderIDsChannel, updatedOrders)
		if err != nil {
			logging.Logger.Error("error occured in order sender")
		}
		close(updatedOrders)
		wg.Done()
	}()

	go func() {
		err := oe.orderUpdater.UpdateOrders(ctx, updatedOrders)
		if err != nil {
			logging.Logger.Error("error occured in order updater")
		}
		wg.Done()
	}()

	wg.Wait()
}

func (oe *OrderEnricher) Run(ctx context.Context) error {
	ticker := time.NewTicker(oe.iterationPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("order enricher: graceful shutdown")
		case <-ticker.C:
			iterationCtx, iterationCtxCancel := context.WithTimeout(ctx, oe.iterationTimeout)
			oe.runIteration(iterationCtx)
			iterationCtxCancel()
		}
	}
}
