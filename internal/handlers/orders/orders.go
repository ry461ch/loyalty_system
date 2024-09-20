package orderhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/services"
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
	if err != nil {
		switch {
		case errors.Is(err, exceptions.NewOrderConflictAnotherUserError()):
			res.WriteHeader(http.StatusConflict)
			return
		case errors.Is(err, exceptions.NewOrderConflictSameUserError()):
			res.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, exceptions.NewOrderBadIDFormatError()):
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		default:
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	res.WriteHeader(http.StatusAccepted)
}

func (oh *OrderHandlers) GetOrders(res http.ResponseWriter, req *http.Request) {
	userID, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := oh.orderService.GetUserOrders(req.Context(), userID)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(orders)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
