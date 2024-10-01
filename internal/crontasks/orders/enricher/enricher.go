package orderenricher

import (
	"context"
	"errors"
	"time"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderEnricher struct {
	orderSender      OrderSender
	orderUpdater     OrderUpdater
	orderGetter      OrderGetter
	iterationTimeout time.Duration
	iterationPeriod  time.Duration
}

func NewOrderEnricher(
	orderGetter OrderGetter,
	orderSender OrderSender,
	orderUpdater OrderUpdater,
	cfg *config.Config,
) *OrderEnricher {
	return &OrderEnricher{
		orderGetter:      orderGetter,
		orderSender:      orderSender,
		orderUpdater:     orderUpdater,
		iterationTimeout: cfg.OrderEnricherTimeout,
		iterationPeriod:  cfg.OrderEnricherPeriod,
	}
}

func (oe *OrderEnricher) runIteration(ctx context.Context) {
	logging.Logger.Infof("Order Enricher: start iteration")

	orderIDsChannel := oe.orderGetter.GetWaitingOrderIdsGenerator(ctx)

	updatedOrders := oe.orderSender.SendOrdersGenerator(ctx, orderIDsChannel)

	oe.orderUpdater.UpdateOrders(ctx, updatedOrders)

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
