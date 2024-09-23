package ordersender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/netaddr"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderSender struct {
	accrualAddr *netaddr.NetAddress
	workersNum  int // RateLimit
	client      *resty.Client
}

func getClient(timeout time.Duration, retries int) *resty.Client {
	return resty.New().
		SetContentLength(true).
		SetRetryCount(retries).
		SetTimeout(timeout).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			// 429
			if resp.StatusCode() == http.StatusTooManyRequests {
				retryAfterStr := resp.Header().Get("Retry-After")
				if retryAfterStr == "" {
					return 0, fmt.Errorf("no Retry-After header came from accrual service")
				}
				retryAfter, err := strconv.Atoi(retryAfterStr)
				if err != nil {
					return 0, fmt.Errorf("bad Retry-After header came from accrual service")
				}
				return time.Duration(retryAfter) * time.Second, nil
			}

			// timeout or 5**
			if resp.IsError() || resp.StatusCode() >= 500 {
				return 0, nil
			}

			// 4**
			return 0, fmt.Errorf("bad Request")
		})
}

func NewOrderSender(cfg *config.Config) *OrderSender {
	return &OrderSender{
		accrualAddr: &cfg.AccuralSystemAddr,
		workersNum:  cfg.OrderSenderRateLimit,
		client:      getClient(cfg.OrderSenderAccrualTimeout, cfg.OrderSenderAccrualRetries),
	}
}

func (os *OrderSender) getOrderFromAccrual(ctx context.Context, orderID string) (*order.Order, error) {
	serverURL := "http://" + os.accrualAddr.Host + ":" + strconv.FormatInt(os.accrualAddr.Port, 10)

	resp, err := os.client.R().SetContext(ctx).Get(serverURL + "/api/orders/" + orderID)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() > 400 && resp.StatusCode() < 500 {
		return nil, fmt.Errorf("bad request, accrual returned %d", resp.StatusCode())
	}

	if resp.StatusCode() > 500 {
		return nil, fmt.Errorf("accrual server unavailable")
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	var updatedOrder order.Order
	err = json.Unmarshal(resp.Body(), &updatedOrder)
	if err != nil {
		return nil, err
	}

	return &updatedOrder, nil
}

func (os *OrderSender) getOrderFromAccrualWorker(ctx context.Context, workerID int, orderIDsChannel <-chan string, updatedOrders chan<- order.Order) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("graceful shutdown worker %d", workerID)
		case orderID := <-orderIDsChannel:
			if orderID == "" {
				return nil
			}
			updatedOrder, err := os.getOrderFromAccrual(ctx, orderID)
			if err != nil {
				logging.Logger.Warnf("Order Sender: exceptions occured for orderID: %s: %v", orderID, err)
				break
			}
			if updatedOrder == nil {
				break
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("graceful shutdown worker %d", workerID)
			case updatedOrders <- *updatedOrder:
				break
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("graceful shutdown worker %d", workerID)
		case <-ticker.C:
		}
	}
}

func (os *OrderSender) GetUpdatedOrders(ctx context.Context, orderIDsChannel <-chan string, updatedOrders chan<- order.Order) error {
	logging.Logger.Infof("Order Sender: init with %d workers", os.workersNum)
	wg := new(errgroup.Group)

	for w := 0; w < os.workersNum; w++ {
		workerID := w
		wg.Go(
			func() error {
				err := os.getOrderFromAccrualWorker(ctx, workerID, orderIDsChannel, updatedOrders)
				if err != nil {
					logging.Logger.Errorf("Order Sender: %v", err)
				}
				return err
			},
		)
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	logging.Logger.Info("Order Sender: successfully handled all orders")
	return nil
}
