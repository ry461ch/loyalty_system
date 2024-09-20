package orderhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"

	"github.com/ry461ch/loyalty_system/internal/models/order"
	"github.com/ry461ch/loyalty_system/internal/services/money"
	"github.com/ry461ch/loyalty_system/internal/services/order"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

func mockRouter(orderHandlers *OrderHandlers) chi.Router {
	router := chi.NewRouter()
	router.Get("/api/user/orders", orderHandlers.GetOrders)
	router.Post("/api/user/orders", orderHandlers.PostOrder)
	return router
}

type outputOrder struct {
	ID       string    `json:"number"`
	Status   int       `json:"status"`
	Accrual  float64   `json:"accrual"`
	UploadAt time.Time `json:"upload_at"`
}

func TestGetOrders(t *testing.T) {
	accrual := float64(500)
	existingUserID := uuid.New()
	existingOrders := []order.Order{
		{
			Accrual:   &accrual,
			Status:    order.PROCESSED,
			ID:        "1115",
			CreatedAt: time.Now(),
		},
		{
			Status:    order.NEW,
			ID:        "1321",
			CreatedAt: time.Now(),
		},
	}

	testCases := []struct {
		testName          string
		inputUserID       uuid.UUID
		expectedOrdersNum int
		expectedCode      int
	}{
		{
			testName:          "successful get orders of existing user",
			inputUserID:       existingUserID,
			expectedOrdersNum: 2,
			expectedCode:      http.StatusOK,
		},
		{
			testName:          "successful get orders of new user",
			inputUserID:       uuid.New(),
			expectedOrdersNum: 0,
			expectedCode:      http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			orderStorage := ordermemstorage.NewOrderMemStorage()
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
			orderService := orderservice.NewOrderService(orderStorage, moneyService)
			handlers := NewOrderHandlers(orderService)
			router := mockRouter(handlers)
			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			for _, existingOrder := range existingOrders {
				orderService.InsertOrder(context.TODO(), existingUserID, existingOrder.ID)
				orderService.UpdateOrder(context.TODO(), &existingOrder)
			}

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", tc.inputUserID.String()).
				Execute(http.MethodGet, srv.URL+"/api/user/orders")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			if tc.expectedOrdersNum == 0 {
				return
			}

			var respOrders []outputOrder
			json.Unmarshal(resp.Body(), &respOrders)
			assert.Equal(t, tc.expectedOrdersNum, len(respOrders), "num of orders not equal")

		})
	}
}

func TestPostOrder(t *testing.T) {
	existingUserID := uuid.New()
	existingOrderID := "1115"
	invalidOrderID := "1111"

	testCases := []struct {
		testName     string
		inputUserID  uuid.UUID
		inputOrderID string
		expectedCode int
	}{
		{
			testName:     "successfully saved new order",
			inputUserID:  existingUserID,
			inputOrderID: "1321",
			expectedCode: http.StatusAccepted,
		},
		{
			testName:     "order was already saved",
			inputUserID:  existingUserID,
			inputOrderID: existingOrderID,
			expectedCode: http.StatusOK,
		},
		{
			testName:     "order was already saved by another user",
			inputUserID:  uuid.New(),
			inputOrderID: existingOrderID,
			expectedCode: http.StatusConflict,
		},
		{
			testName:     "invalid order id",
			inputUserID:  uuid.New(),
			inputOrderID: invalidOrderID,
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			orderStorage := ordermemstorage.NewOrderMemStorage()
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
			orderService := orderservice.NewOrderService(orderStorage, moneyService)
			handlers := NewOrderHandlers(orderService)
			router := mockRouter(handlers)
			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			orderService.InsertOrder(context.TODO(), existingUserID, existingOrderID)

			req := []byte(tc.inputOrderID)
			resp, _ := client.R().
				SetHeader("X-User-Id", tc.inputUserID.String()).
				SetBody(req).
				Execute(http.MethodPost, srv.URL+"/api/user/orders")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}
