package orderenricher

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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/components/orders/getter"
	"github.com/ry461ch/loyalty_system/internal/components/orders/sender"
	"github.com/ry461ch/loyalty_system/internal/components/orders/updater"
	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/netaddr"
	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/services/money"
	"github.com/ry461ch/loyalty_system/internal/services/order"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

type MockServerStorage struct {
	timesCalled int
}

type outputOrder struct {
	ID      string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (m *MockServerStorage) handler(res http.ResponseWriter, req *http.Request) {
	m.timesCalled += 1

	orderID := chi.URLParam(req, "order_id")

	var resultOrder *outputOrder
	switch orderID {
	case "1115":
		resultOrder = &outputOrder{
			ID:      orderID,
			Status:  "PROCESSED",
			Accrual: 100,
		}
	case "1214":
		resultOrder = &outputOrder{
			ID:     orderID,
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

func TestEnricher(t *testing.T) {
	serverStorage := MockServerStorage{}
	router := serverStorage.mockRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	accrual := float64(100)
	existingUserID := uuid.New()
	existingOrders := []order.Order{
		{
			ID:     "1115",
			Status: order.NEW,
		},
		{
			ID:     "1321",
			Status: order.INVALID,
		},
		{
			ID:     "1214",
			Status: order.PROCESSING,
		},
		{
			ID:     "1313",
			Status: order.NEW,
		},
	}
	expectedOrders := []order.Order{
		{
			ID:     "1313",
			Status: order.NEW,
		},
		{
			ID:     "1321",
			Status: order.INVALID,
		},
		{
			ID:      "1115",
			Status:  order.PROCESSED,
			Accrual: &accrual,
		},
		{
			ID:     "1214",
			Status: order.INVALID,
		},
	}

	orderStorage := ordermemstorage.NewOrderMemStorage()
	balanceStorage := balancememstorage.NewBalanceMemStorage()
	withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
	for _, existingOrder := range existingOrders {
		orderStorage.InsertOrder(context.TODO(), existingUserID, existingOrder.ID, nil)
		if existingOrder.Status != order.NEW {
			orderStorage.UpdateOrder(context.TODO(), &existingOrder, nil)
		}
	}
	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	orderService := orderservice.NewOrderService(orderStorage, moneyService)

	cfg := config.Config{
		AccuralSystemAddr:         *splitURL(srv.URL),
		OrderUpdaterRateLimit:     10,
		OrderGetterOrdersLimit:    2,
		OrderGetterRateLimit:      1,
		OrderSenderRateLimit:      2,
		OrderSenderAccrualTimeout: time.Millisecond * 500,
		OrderSenderAccrualRetries: 3,
		OrderEnricherChannelSize:  10,
	}

	sender := ordersender.NewOrderSender(&cfg)
	updater := orderupdater.NewOrderUpdater(orderService, &cfg)
	getter := ordergetter.NewOrderGetter(orderService, &cfg)

	enricher := NewOrderEnricher(getter, sender, updater, &cfg)

	start := time.Now().UTC()
	enricher.runIteration(context.TODO())
	assert.GreaterOrEqual(t, time.Since(start), time.Second*2, "workers worked less than 2 seconds")

	updatedOrdersList, _ := orderStorage.GetUserOrders(context.TODO(), existingUserID)

	updatedOrders := map[string]order.Order{}
	for _, updatedOrder := range updatedOrdersList {
		updatedOrders[updatedOrder.ID] = updatedOrder
	}

	for _, expectedOrder := range expectedOrders {
		updatedOrder, ok := updatedOrders[expectedOrder.ID]
		assert.True(t, ok, "orser was not in updated list")
		assert.Equal(t, updatedOrder.Status, expectedOrder.Status, "statuses not equal")
		if expectedOrder.Accrual != nil {
			assert.Equal(t, expectedOrder.Accrual, updatedOrder.Accrual, "accrual nor equal")
		}
	}

	expectedBalance := balance.Balance{
		Current:   accrual,
		Withdrawn: 0,
	}
	userBalance, _ := balanceStorage.GetBalance(context.TODO(), existingUserID)
	assert.Equal(t, expectedBalance, *userBalance, "balances not equal")
}
