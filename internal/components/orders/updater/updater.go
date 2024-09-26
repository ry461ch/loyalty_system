package orderupdater

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderUpdater struct {
	orderService OrderUpdaterService
	workersNum   int // RateLimit for SQL updating query
}

func NewOrderUpdater(orderService OrderUpdaterService, cfg *config.Config) *OrderUpdater {
	return &OrderUpdater{
		orderService: orderService,
		workersNum:   cfg.OrderUpdaterRateLimit,
	}
}

func (ou *OrderUpdater) updateOrderWorker(ctx context.Context, workerID int, updatedOrders <-chan order.Order) error {
	for updatedOrder := range updatedOrders {
		select {
		case <-ctx.Done():
			return fmt.Errorf("worker %d %w", workerID, exceptions.ErrGracefullyShutDown)
		default:
		}

		if updatedOrder.ID == "" {
			// if channel closed
			return nil
		}

		err := ou.orderService.UpdateOrder(ctx, &updatedOrder)
		if err != nil {
			logging.Logger.Warnf("Order Updater: exceptions occured for orderID: %s: %s", updatedOrder.ID, err.Error())
			return nil
		}

		time.Sleep(time.Second)
	}
	return nil
}

func (ou *OrderUpdater) UpdateOrders(ctx context.Context, updatedOrders <-chan order.Order) {
	logging.Logger.Infof("Order Updater: init with %d workers", ou.workersNum)
	var wg sync.WaitGroup
	wg.Add(ou.workersNum)

	for w := 0; w < ou.workersNum; w++ {
		workerID := w
		go func() {
			err := ou.updateOrderWorker(ctx, workerID, updatedOrders)
			if err != nil {
				if errors.Is(err, exceptions.ErrGracefullyShutDown) {
					logging.Logger.Infof("Order Updater: worker %d gracefully shutdown", workerID)
					wg.Done()
					return
				}
				logging.Logger.Errorf("Order Updater: %v", err)
				wg.Done()
				return
			}
			logging.Logger.Infof("Order Updater:  worker %d successfully ended his work", workerID)
			wg.Done()
		}()
	}

	wg.Wait()

	logging.Logger.Info("Order Updater: gracefully shutdown")
}
