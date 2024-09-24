package orderupdater

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/services"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderUpdater struct {
	orderService services.OrderService
	workersNum   int // RateLimit for SQL updating query
}

func NewOrderUpdater(orderService services.OrderService, cfg *config.Config) *OrderUpdater {
	return &OrderUpdater{
		orderService: orderService,
		workersNum:   cfg.OrderUpdaterRateLimit,
	}
}

func (ou *OrderUpdater) updateOrderWorker(ctx context.Context, workerID int, updatedOrders <-chan order.Order) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("worker %d %w", workerID, exceptions.ErrGracefullyShutDown)
		case updatedOrder := <-updatedOrders:
			if updatedOrder.ID == "" {
				return nil
			}

			err := ou.orderService.UpdateOrder(ctx, &updatedOrder)
			if err != nil {
				logging.Logger.Warnf("Order Updater: exceptions occured for orderID: %s: %s", updatedOrder.ID, err.Error())
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("worker %d %w", workerID, exceptions.ErrGracefullyShutDown)
		case <-ticker.C:
		}
	}
}

func (ou *OrderUpdater) UpdateOrders(ctx context.Context, updatedOrders <-chan order.Order) error {
	logging.Logger.Infof("Order Updater: init with %d workers", ou.workersNum)
	wg := new(errgroup.Group)

	for w := 0; w < ou.workersNum; w++ {
		workerID := w
		wg.Go(
			func() error {
				err := ou.updateOrderWorker(ctx, workerID, updatedOrders)
				if err != nil {
					if errors.Is(err, exceptions.ErrGracefullyShutDown) {
						logging.Logger.Infof("Order Updater: worker %d gracefully shutdown", workerID)
						return nil
					}
					logging.Logger.Errorf("Order Updater: %v", err)
				}
				return err
			},
		)
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	logging.Logger.Info("Order Updater: gracefully shutdown")
	return nil
}
