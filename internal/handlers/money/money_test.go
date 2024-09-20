package moneyhandlers

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

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/services/money"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

func mockRouter(moneyHandlers *MoneyHandlers) chi.Router {
	router := chi.NewRouter()
	router.Get("/api/user/balance", moneyHandlers.GetBalance)
	router.Post("/api/user/balance/withdraw", moneyHandlers.PostWithdrawal)
	router.Get("/api/user/withdrawals", moneyHandlers.GetWithdrawals)
	return router
}

func TestGetBalance(t *testing.T) {
	existingUserID := uuid.New()
	existingBalance := balance.Balance{
		Current:   200,
		Withdrawn: 300,
	}
	existingWithdrawalID := uuid.New()
	createdAt := time.Now()
	existingWithdrawal := withdrawal.Withdrawal{
		ID:        &existingWithdrawalID,
		OrderID:   "1115",
		UserID:    &existingUserID,
		Sum:       existingBalance.Withdrawn,
		CreatedAt: &createdAt,
	}

	balanceStorage := balancememstorage.NewBalanceMemStorage()
	withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	handlers := NewMoneyHandlers(moneyService)
	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	testCases := []struct {
		testName        string
		inputUserID     uuid.UUID
		expectedBalance balance.Balance
		expectedCode    int
	}{
		{
			testName:        "successful get balance of existing user",
			inputUserID:     existingUserID,
			expectedBalance: existingBalance,
			expectedCode:    http.StatusOK,
		},
		{
			testName:        "successful get balance of new user",
			inputUserID:     uuid.New(),
			expectedBalance: balance.Balance{},
			expectedCode:    http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			moneyService.AddAccrual(context.TODO(), existingUserID, existingBalance.Current+existingBalance.Withdrawn, nil)
			moneyService.Withdraw(context.TODO(), &existingWithdrawal)

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", tc.inputUserID.String()).
				Execute(http.MethodGet, srv.URL+"/api/user/balance")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			var respBalance balance.Balance
			json.Unmarshal(resp.Body(), &respBalance)
			assert.Equal(t, tc.expectedBalance, respBalance, "balances not equal")
		})
	}
}

type outputWithdrawal struct {
	OrderID     string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func TestGetWithdrawals(t *testing.T) {
	existingUserID := uuid.New()
	existingWithdrawalID1 := uuid.New()
	existingWithdrawalID2 := uuid.New()
	existingWithdrawal1 := withdrawal.Withdrawal{
		ID:      &existingWithdrawalID1,
		OrderID: "1115",
		UserID:  &existingUserID,
		Sum:     300,
	}
	existingWithdrawal2 := withdrawal.Withdrawal{
		ID:      &existingWithdrawalID2,
		UserID:  &existingUserID,
		OrderID: "1321",
		Sum:     200,
	}

	balanceStorage := balancememstorage.NewBalanceMemStorage()
	withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
	moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
	handlers := NewMoneyHandlers(moneyService)
	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	testCases := []struct {
		testName               string
		inputUserID            uuid.UUID
		expectedWithdrawalsNum int
		expectedCode           int
	}{
		{
			testName:               "successful get withdrawals of existing user",
			inputUserID:            existingUserID,
			expectedWithdrawalsNum: 2,
			expectedCode:           http.StatusOK,
		},
		{
			testName:               "successful get withdrawals of new user",
			inputUserID:            uuid.New(),
			expectedWithdrawalsNum: 0,
			expectedCode:           http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			moneyService.AddAccrual(context.TODO(), existingUserID, existingWithdrawal1.Sum+existingWithdrawal2.Sum, nil)
			err := moneyService.Withdraw(context.TODO(), &existingWithdrawal1)
			assert.Nil(t, err)
			err = moneyService.Withdraw(context.TODO(), &existingWithdrawal2)
			assert.Nil(t, err)

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", tc.inputUserID.String()).
				Execute(http.MethodGet, srv.URL+"/api/user/withdrawals")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			if tc.expectedWithdrawalsNum != 0 {
				var respWithdrawals []outputWithdrawal
				json.Unmarshal(resp.Body(), &respWithdrawals)
				assert.Equal(t, tc.expectedWithdrawalsNum, len(respWithdrawals), "withdrawals not equal")
			}
		})
	}
}

type InputWithdrawal struct {
	OrderID string  `json:"order"`
	Sum     float64 `json:"sum"`
}

func TestPostWithdraw(t *testing.T) {
	existingUserID := uuid.New()
	existingBalanceCurrent := float64(200)
	existingWithdrawalID := uuid.New()
	existingWithdrawal := withdrawal.Withdrawal{
		ID:      &existingWithdrawalID,
		OrderID: "1115",
		UserID:  &existingUserID,
		Sum:     300,
	}

	testCases := []struct {
		testName               string
		inputIdempotencyToken  string
		inputWithdrawal        *InputWithdrawal
		expectedWithdrawalsNum int
		expectedCode           int
	}{
		{
			testName:              "successful withdraw of existing user",
			inputIdempotencyToken: uuid.NewString(),
			inputWithdrawal: &InputWithdrawal{
				OrderID: "1321",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 2,
			expectedCode:           http.StatusOK,
		},
		{
			testName:               "empty input",
			inputIdempotencyToken:  uuid.NewString(),
			inputWithdrawal:        nil,
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusBadRequest,
		},
		{
			testName:              "invalid withdrawal sum",
			inputIdempotencyToken: uuid.NewString(),
			inputWithdrawal: &InputWithdrawal{
				OrderID: "1321",
				Sum:     0,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusBadRequest,
		},
		{
			testName:              "not enough money on balance",
			inputIdempotencyToken: uuid.NewString(),
			inputWithdrawal: &InputWithdrawal{
				OrderID: "1321",
				Sum:     existingBalanceCurrent + 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusPaymentRequired,
		},
		{
			testName:              "invalid order id format",
			inputIdempotencyToken: uuid.NewString(),
			inputWithdrawal: &InputWithdrawal{
				OrderID: "1322",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusUnprocessableEntity,
		},
		{
			testName:              "existing idempotency token",
			inputIdempotencyToken: existingWithdrawal.ID.String(),
			inputWithdrawal: &InputWithdrawal{
				OrderID: "1321",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusOK,
		},
		{
			testName:              "invalid idempotency token",
			inputIdempotencyToken: "invalid",
			inputWithdrawal: &InputWithdrawal{
				OrderID: "1321",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusBadRequest,
		},
		{
			testName:              "empty idempotency token",
			inputIdempotencyToken: "",
			inputWithdrawal: &InputWithdrawal{
				OrderID: "1321",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 2,
			expectedCode:           http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			moneyService := moneyservice.NewMoneyService(balanceStorage, withdrawalStorage)
			handlers := NewMoneyHandlers(moneyService)
			router := mockRouter(handlers)
			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			moneyService.AddAccrual(context.TODO(), existingUserID, existingBalanceCurrent+existingWithdrawal.Sum, nil)
			moneyService.Withdraw(context.TODO(), &existingWithdrawal)

			var req []byte
			if tc.inputWithdrawal != nil {
				req, _ = json.Marshal(*tc.inputWithdrawal)
			} else {
				req, _ = json.Marshal(map[string]string{})
			}

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", existingUserID.String()).
				SetHeader("Idempotency-Key", tc.inputIdempotencyToken).
				SetBody(req).
				Execute(http.MethodPost, srv.URL+"/api/user/balance/withdraw")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			userWithdrawals, _ := moneyService.GetWithdrawals(context.TODO(), existingUserID)
			assert.Equal(t, tc.expectedWithdrawalsNum, len(userWithdrawals), "num of withdrawals don't match")
		})
	}
}
