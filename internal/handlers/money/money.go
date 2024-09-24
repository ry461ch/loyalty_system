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
	"github.com/ry461ch/loyalty_system/pkg/logging"
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
	userID, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		logging.Logger.Errorf("Withdraw: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var withdrawalID uuid.UUID
	withdrawalIDStr := req.Header.Get("Idempotency-Key")
	if withdrawalIDStr == "" {
		withdrawalID = uuid.New()
	} else {
		withdrawalID, err = uuid.Parse(withdrawalIDStr)
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
	inputWithdrawal.ID = &withdrawalID
	inputWithdrawal.UserID = &userID

	err = mh.moneyService.Withdraw(req.Context(), &inputWithdrawal)
	if err != nil {
		switch {
		case errors.Is(err, exceptions.ErrNotEnoughBalance):
			res.WriteHeader(http.StatusPaymentRequired)
			return
		case errors.Is(err, exceptions.ErrOrderBadIDFormat):
			res.WriteHeader(http.StatusUnprocessableEntity)
			return
		default:
			logging.Logger.Errorf("Withdraw: internal error: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	res.WriteHeader(http.StatusOK)
}

func (mh *MoneyHandlers) GetWithdrawals(res http.ResponseWriter, req *http.Request) {
	userID, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		logging.Logger.Errorf("Get withdrawals: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	userWithdrawals, err := mh.moneyService.GetWithdrawals(req.Context(), userID)
	if err != nil {
		logging.Logger.Errorf("Get withdrawals: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(userWithdrawals) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(userWithdrawals)
	if err != nil {
		logging.Logger.Errorf("Get withdrawals: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

func (mh *MoneyHandlers) GetBalance(res http.ResponseWriter, req *http.Request) {
	userID, err := uuid.Parse(req.Header.Get("X-User-Id"))
	if err != nil {
		logging.Logger.Errorf("Get balance: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	userBalance, err := mh.moneyService.GetBalance(req.Context(), userID)
	if err != nil {
		logging.Logger.Errorf("Get balance: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(userBalance)
	if err != nil {
		logging.Logger.Errorf("Get balance: internal error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
