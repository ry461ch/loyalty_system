package orderenricher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ry461ch/loyalty_system/internal/components/orders"
	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderEnricher struct {
	orderSender          ordercomponents.OrderSender
	orderUpdater         ordercomponents.OrderUpdater
	orderGetter          ordercomponents.OrderGetter
	iterationChannelSize int
	iterationTimeout     time.Duration
	iterationPeriod      time.Duration
}

func NewOrderEnricher(
	orderGetter ordercomponents.OrderGetter,
	orderSender ordercomponents.OrderSender,
	orderUpdater ordercomponents.OrderUpdater,
	cfg *config.Config,
) *OrderEnricher {
	return &OrderEnricher{
		orderGetter:          orderGetter,
		orderSender:          orderSender,
		orderUpdater:         orderUpdater,
		iterationChannelSize: cfg.OrderEnricherChannelSize,
		iterationTimeout:     cfg.OrderEnricherTimeout,
		iterationPeriod:      cfg.OrderEnricherPeriod,
	}
}

func (oe *OrderEnricher) runIteration(ctx context.Context) {
	logging.Logger.Infof("Order Enricher: start iteration")
	orderIDsChannel := make(chan string, oe.iterationChannelSize)
	updatedOrders := make(chan order.Order, oe.iterationChannelSize)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		err := oe.orderGetter.GetWaitingOrderIDs(ctx, orderIDsChannel)
		if err != nil {
			logging.Logger.Errorf("Order Enricher: error occured in order getter: %v", err)
		}
		close(orderIDsChannel)
		wg.Done()
	}()

	go func() {
		err := oe.orderSender.GetUpdatedOrders(ctx, orderIDsChannel, updatedOrders)
		if err != nil {
			logging.Logger.Errorf("Order Enricher: error occured in order sender: %v", err)
		}
		close(updatedOrders)
		wg.Done()
	}()

	go func() {
		err := oe.orderUpdater.UpdateOrders(ctx, updatedOrders)
		if err != nil {
			logging.Logger.Errorf("Order Enricher: error occured in order updater: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()
	logging.Logger.Infof("Order Enricher: end iteration")
}

func (oe *OrderEnricher) Run(ctx context.Context) error {
	logging.Logger.Infof("Order Enricher: started")
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
