package orderhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/interfaces/services"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type OrderHandlers struct {
	orderService services.OrderService
}

func NewOrderHandlers(orderService services.OrderService) *OrderHandlers {
	return &OrderHandlers{
		orderService: orderService,
	}
}

func (oh *OrderHandlers) PostOrder(res http.ResponseWriter, req *http.Request) {
	userID, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		logging.Logger.Errorf("New order: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID := string(reqBody)

	err = oh.orderService.InsertOrder(req.Context(), userID, orderID)
	if err == nil {
		res.WriteHeader(http.StatusAccepted)
		return
	}

	switch {
	case errors.Is(err, exceptions.ErrOrderConflictAnotherUser):
		res.WriteHeader(http.StatusConflict)
	case errors.Is(err, exceptions.ErrOrderConflictSameUser):
		res.WriteHeader(http.StatusOK)
	case errors.Is(err, exceptions.ErrOrderBadIDFormat):
		res.WriteHeader(http.StatusUnprocessableEntity)
	default:
		logging.Logger.Errorf("New order: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (oh *OrderHandlers) GetOrders(res http.ResponseWriter, req *http.Request) {
	userID, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := oh.orderService.GetUserOrders(req.Context(), userID)
	if err != nil {
		logging.Logger.Errorf("Get orders: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(orders)
	if err != nil {
		logging.Logger.Errorf("Get orders: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
