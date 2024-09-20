package moneyhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/services"
)

type MoneyHandlers struct {
	moneyService services.MoneyService
}

func NewMoneyHandlers(moneyService services.MoneyService) *MoneyHandlers {
	return &MoneyHandlers{
		moneyService: moneyService,
	}
}

func (mh *MoneyHandlers) PostWithdrawal(res http.ResponseWriter, req *http.Request) {
	userId, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var withdrawalId uuid.UUID
	withdrawalIdStr := req.Header.Get("Idempotency-Key")
	if withdrawalIdStr == "" {
		withdrawalId = uuid.New()
	} else {
		withdrawalId, err = uuid.Parse(withdrawalIdStr)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var inputWithdrawal withdrawal.Withdrawal
	err = json.Unmarshal(reqBody, &inputWithdrawal)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	inputWithdrawal.Id = &withdrawalId
	inputWithdrawal.UserId = &userId

	err = mh.moneyService.Withdraw(req.Context(), &inputWithdrawal)
	if err != nil {
		switch {
		case errors.Is(err, exceptions.NewBalanceNotEnoughBalanceError()):
			res.WriteHeader(http.StatusPaymentRequired)
			return
		case errors.Is(err, exceptions.NewOrderBadIdFormatError()):
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		default:
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	res.WriteHeader(http.StatusOK)
}

func (mh *MoneyHandlers) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	userId, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	userWithdrawals, err := mh.moneyService.GetWithdrawals(req.Context(), userId)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(userWithdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(userWithdrawals)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

func (mh *MoneyHandlers) GetBalance(res http.ResponseWriter, req *http.Request) {
	userId, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	userBalance, err := mh.moneyService.GetBalance(req.Context(), userId)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(userBalance)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
