package ordersender

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/netaddr"
	"github.com/ry461ch/loyalty_system/internal/models/order"
)

type MockServerStorage struct {
	timesCalled    int
}

type outputOrder struct {
	ID string  `json:"order"`
	Status string `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (m *MockServerStorage) handler(res http.ResponseWriter, req *http.Request) {
	m.timesCalled += 1

	orderID := chi.URLParam(req, "order_id")
	
	var resultOrder *outputOrder
	switch orderID {
	case "1115":
		resultOrder = &outputOrder{
			ID: orderID,
			Status: "PROCESSED",
			Accrual: 100,
		}
	case "1214":
		resultOrder = &outputOrder{
			ID: orderID,
			Status: "INVALID",
		}
	}

	if resultOrder == nil {
		res.WriteHeader(http.StatusNoContent)
	}

	resp, _ := json.Marshal(resultOrder)
	res.Write(resp)
}

func (m *MockServerStorage) mockRouter() chi.Router {
	router := chi.NewRouter()
	router.Post("/api/orders/{order_id:[0-9]+}", m.handler)
	return router
}

func splitURL(URL string) *netaddr.NetAddress {
	updatedURL, _ := strings.CutPrefix(URL, "http://")
	parts := strings.Split(updatedURL, ":")
	port, _ := strconv.ParseInt(parts[1], 10, 0)
	return &netaddr.NetAddress{Host: parts[0], Port: port}
}

func TestSender(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	orderIDsChannel := make(chan string, 3)
	orderIDsChannel <- "1313"
	orderIDsChannel <- "1115"
	orderIDsChannel <- "1214"
	close(orderIDsChannel)

	updatedOrdersChannel := make(chan order.Order, 2)
	defer close(updatedOrdersChannel)

	sender := OrderSender{
		accrualAddr: splitURL(srv.URL), 
		workersNum: 2,
		client: getClient(time.Millisecond * 500, 3),
	}

	start := time.Now()
	sender.GetUpdatedOrders(context.TODO(), orderIDsChannel, updatedOrdersChannel)
	assert.GreaterOrEqual(t, time.Since(start), time.Second * 2, "workers worked less than 2 seconds")

	assert.Equal(t, 3, serverStorage.timesCalled, "Не прошел запрос на сервер")

	order1 := <-updatedOrdersChannel
	order2 := <-updatedOrdersChannel

	updatedOrders := map[string]order.Order{
		order1.ID: order1, order2.ID: order2,
	}

	accrual := float64(100)
	expectedOrders := []order.Order{
		{
			ID: "1115",
			Status: order.PROCESSED,
			Accrual: &accrual,
		},
		{
			ID: "1214",
			Status: order.INVALID,
		},
	}
	
	for _, expectedOrder := range expectedOrders {
		updatedOrder, ok := updatedOrders[expectedOrder.ID]
		assert.True(t, ok, "orser was not in updated list")
		assert.Equal(t, updatedOrder.Status, expectedOrder.Status, "statuses not equal")
		if expectedOrder.Accrual != nil {
			assert.Equal(t, expectedOrder.Accrual, updatedOrder.Accrual, "accrual nor equal")
		}
	}
}
