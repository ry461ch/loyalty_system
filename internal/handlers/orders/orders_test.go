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
	"github.com/ry461ch/loyalty_system/internal/storage/memory"
)

func mockRouter(orderHandlers *OrderHandlers) chi.Router {
	router := chi.NewRouter()
	router.Get("/api/user/orders", orderHandlers.GetOrders)
	router.Post("/api/user/orders", orderHandlers.PostOrder)
	return router
}

func TestGetOrders(t *testing.T) {
	accrual := float64(500)
	existingUserId := uuid.New()
	existingOrders := []order.Order{
		{
			Accrual:   &accrual,
			Status:    order.PROCESSED,
			Id:        "1115",
			CreatedAt: time.Now(),
		},
		{
			Status:    order.NEW,
			Id:        "1321",
			CreatedAt: time.Now(),
		},
	}

	testCases := []struct {
		testName               string
		inputUserId            uuid.UUID
		expectedOrdersNum      int
		expectedCode           int
	}{
		{
			testName:               "successful get orders of existing user",
			inputUserId:            existingUserId,
			expectedOrdersNum: 2,
			expectedCode:           http.StatusOK,
		},
		{
			testName:               "successful get orders of new user",
			inputUserId:            uuid.New(),
			expectedOrdersNum: 0,
			expectedCode:           http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := memstorage.NewMemStorage()
			moneyService := moneyservice.NewMoneyService(storage)
			orderService := orderservice.NewOrderService(storage, moneyService)
			handlers := NewOrderHandlers(orderService)
			router := mockRouter(handlers)
			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			for _, existingOrder := range(existingOrders) {
				orderService.InsertOrder(context.TODO(), existingUserId, existingOrder.Id)
				orderService.UpdateOrder(context.TODO(), &existingOrder)
			}

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", tc.inputUserId.String()).
				Execute(http.MethodGet, srv.URL+"/api/user/orders")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			if tc.expectedOrdersNum != 0 {
				var respOrders []order.Order
				json.Unmarshal(resp.Body(), &respOrders)
				assert.Equal(t, tc.expectedOrdersNum, len(respOrders), "num of orders not equal")
			}
		})
	}
}

func TestPostOrder(t *testing.T) {
	existingUserId := uuid.New()
	existingOrderId := "1115"
	invalidOrderid := "1111"

	testCases := []struct {
		testName               string
		inputUserId            uuid.UUID
		inputOrderId		   string
		expectedCode           int
	}{
		{
			testName:               "successfully saved new order",
			inputUserId:            existingUserId,
			inputOrderId:			"1321",
			expectedCode:           http.StatusAccepted,
		},
		{
			testName:               "order was already saved",
			inputUserId:            existingUserId,
			inputOrderId:			existingOrderId,
			expectedCode:           http.StatusOK,
		},
		{
			testName:               "order was already saved by another user",
			inputUserId:            uuid.New(),
			inputOrderId:			existingOrderId,
			expectedCode:           http.StatusConflict,
		},
		{
			testName:               "invalid order id",
			inputUserId:            uuid.New(),
			inputOrderId:			invalidOrderid,
			expectedCode:           http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := memstorage.NewMemStorage()
			moneyService := moneyservice.NewMoneyService(storage)
			orderService := orderservice.NewOrderService(storage, moneyService)
			handlers := NewOrderHandlers(orderService)
			router := mockRouter(handlers)
			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			orderService.InsertOrder(context.TODO(), existingUserId, existingOrderId)

			req := []byte(tc.inputOrderId)
			resp, _ := client.R().
				SetHeader("X-User-Id", tc.inputUserId.String()).
				SetBody(req).
				Execute(http.MethodPost, srv.URL+"/api/user/orders")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
		})
	}
}
