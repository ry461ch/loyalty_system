package ordersender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/netaddr"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderSender struct {
	accrualAddr *netaddr.NetAddress
	workersNum  int // RateLimit
	channelSize int
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
		channelSize: cfg.OrderSenderChannelSize,
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
	for orderID := range orderIDsChannel {
		select {
		case <-ctx.Done():
			return fmt.Errorf("worker %d %w", workerID, exceptions.ErrGracefullyShutDown)
		default:
		}

		updatedOrder, err := os.getOrderFromAccrual(ctx, orderID)
		if err != nil {
			logging.Logger.Warnf("Order Sender: exceptions occured for orderID: %s: %v", orderID, err)
		}
		if updatedOrder != nil {
			updatedOrders <- *updatedOrder
		}

		time.Sleep(time.Second)
	}
	return nil
}

func (os *OrderSender) sendOrders(ctx context.Context, orderIDsChannel <-chan string, updatedOrders chan<- order.Order) {
	logging.Logger.Infof("Order Sender: init with %d workers", os.workersNum)
	var wg sync.WaitGroup
	wg.Add(os.workersNum)

	for w := 0; w < os.workersNum; w++ {
		workerID := w
		go func() {
			err := os.getOrderFromAccrualWorker(ctx, workerID, orderIDsChannel, updatedOrders)
			if err != nil {
				if errors.Is(err, exceptions.ErrGracefullyShutDown) {
					logging.Logger.Infof("Order Sender:  worker %d gracefully shutdown", workerID)
					wg.Done()
					return
				}
				logging.Logger.Errorf("Order Sender: %v", err)
				wg.Done()
				return
			}
			logging.Logger.Infof("Order Sender: worker %d successfully ended his work", workerID)
			wg.Done()
		}()
	}

	wg.Wait()
	logging.Logger.Info("Order Sender: gracefully shutdown")
}

func (os *OrderSender) SendOrdersGenerator(ctx context.Context, orderIDsChannel <-chan string) chan order.Order {
	updatedOrders := make(chan order.Order, os.channelSize)

	go func() {
		defer close(updatedOrders)
		os.sendOrders(ctx, orderIDsChannel, updatedOrders)
	}()

	return updatedOrders
}
