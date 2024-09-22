package ordersender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
	"github.com/go-resty/resty/v2"

	"github.com/ry461ch/loyalty_system/internal/models/netaddr"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderSender struct {
	accrualAddr *netaddr.NetAddress
	workersNum  int  // RateLimit
	client *resty.Client
}

func NewOrderSender(accrualAddr *netaddr.NetAddress, workersNum int) *OrderSender {
	return &OrderSender{
		accrualAddr: accrualAddr,
		workersNum: workersNum,
		client: resty.New().
			SetContentLength(true).
			SetRetryCount(3).
			SetTimeout(500 * time.Millisecond).
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
			}),
	}
}

func (os *OrderSender) getOrderFromAccrual(ctx context.Context, orderID string) (*order.Order, error) {
	serverURL := "http://" + os.accrualAddr.Host + ":" + strconv.FormatInt(os.accrualAddr.Port, 10)

	resp, err := os.client.R().SetContext(ctx).Post(serverURL + "/api/orders/" + orderID)
	if err != nil {
		return nil, err
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
			return fmt.Errorf("order sender: graceful shutdown worker %d", workerID)
		case orderID := <-orderIDsChannel:
			if orderID == "" {
				return nil
			}
			updatedOrder, err := os.getOrderFromAccrual(ctx, orderID)
			if err != nil {
				logging.Logger.Warnf("OrderSender: exceptions occured for orderID: %s: %s", orderID, err.Error())
				break
			}
			if updatedOrder == nil {
				break
			}

			select {
				case <-ctx.Done():
					return fmt.Errorf("order sender: graceful shutdown worker %d", workerID)
				case updatedOrders <- *updatedOrder:
					break
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("order sender: graceful shutdown worker %d", workerID)
		case <-ticker.C:
		}
	}
}

func (os *OrderSender) GetUpdatedOrders(ctx context.Context, orderIDsChannel <-chan string, updatedOrders chan<- order.Order) error {
	wg := new(errgroup.Group)

	for w := 0; w < os.workersNum; w++ {
		workerID := w
		wg.Go(
			func() error {
				return os.getOrderFromAccrualWorker(ctx, workerID, orderIDsChannel, updatedOrders)
			},
		)
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	return nil
}
