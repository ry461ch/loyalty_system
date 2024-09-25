package orderenricher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/interfaces/components"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderEnricher struct {
	orderSender          components.OrderSender
	orderUpdater         components.OrderUpdater
	orderGetter          components.OrderGetter
	iterationChannelSize int
	iterationTimeout     time.Duration
	iterationPeriod      time.Duration
}

func NewOrderEnricher(
	orderGetter components.OrderGetter,
	orderSender components.OrderSender,
	orderUpdater components.OrderUpdater,
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
		oe.orderGetter.GetWaitingOrderIDs(ctx, orderIDsChannel)
		close(orderIDsChannel)
		wg.Done()
	}()

	go func() {
		oe.orderSender.GetUpdatedOrders(ctx, orderIDsChannel, updatedOrders)
		close(updatedOrders)
		wg.Done()
	}()

	go func() {
		oe.orderUpdater.UpdateOrders(ctx, updatedOrders)
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
