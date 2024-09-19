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
	"github.com/ry461ch/loyalty_system/internal/storage/memory"
)

func mockRouter(moneyHandlers *MoneyHandlers) chi.Router {
	router := chi.NewRouter()
	router.Get("/api/user/balance", moneyHandlers.GetBalance)
	router.Post("/api/user/balance/withdraw", moneyHandlers.PostWithdrawal)
	router.Get("/api/user/withdrawals", moneyHandlers.GetWithdrawals)
	return router
}

func TestGetBalance(t *testing.T) {
	existingUserId := uuid.New()
	existingBalance := balance.Balance{
		Current:   200,
		Withdrawn: 300,
	}
	existingWithdrawal := withdrawal.Withdrawal{
		Id:        uuid.New(),
		OrderId:   "1115",
		Sum:       existingBalance.Withdrawn,
		CreatedAt: time.Now(),
	}

	storage := memstorage.NewMemStorage()
	moneyService := moneyservice.NewMoneyService(storage)
	handlers := NewMoneyHandlers(moneyService)
	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	testCases := []struct {
		testName        string
		inputUserId     uuid.UUID
		expectedBalance balance.Balance
		expectedCode    int
	}{
		{
			testName:        "successful get balance of existing user",
			inputUserId:     existingUserId,
			expectedBalance: existingBalance,
			expectedCode:    http.StatusOK,
		},
		{
			testName:        "successful get balance of new user",
			inputUserId:     uuid.New(),
			expectedBalance: balance.Balance{},
			expectedCode:    http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			moneyService.AddAccrual(context.TODO(), existingUserId, existingBalance.Current+existingBalance.Withdrawn, nil)
			moneyService.Withdraw(context.TODO(), existingUserId, &existingWithdrawal)

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", tc.inputUserId.String()).
				Execute(http.MethodGet, srv.URL+"/api/user/balance")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			var respBalance balance.Balance
			json.Unmarshal(resp.Body(), &respBalance)
			assert.Equal(t, tc.expectedBalance, respBalance, "balances not equal")
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	existingUserId := uuid.New()
	existingWithdrawal1 := withdrawal.Withdrawal{
		Id:      uuid.New(),
		OrderId: "1115",
		Sum:     300,
	}
	existingWithdrawal2 := withdrawal.Withdrawal{
		Id:      uuid.New(),
		OrderId: "1321",
		Sum:     200,
	}

	storage := memstorage.NewMemStorage()
	moneyService := moneyservice.NewMoneyService(storage)
	handlers := NewMoneyHandlers(moneyService)
	router := mockRouter(handlers)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := resty.New()

	testCases := []struct {
		testName               string
		inputUserId            uuid.UUID
		expectedWithdrawalsNum int
		expectedCode           int
	}{
		{
			testName:               "successful get withdrawals of existing user",
			inputUserId:            existingUserId,
			expectedWithdrawalsNum: 2,
			expectedCode:           http.StatusOK,
		},
		{
			testName:               "successful get withdrawals of new user",
			inputUserId:            uuid.New(),
			expectedWithdrawalsNum: 0,
			expectedCode:           http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			moneyService.AddAccrual(context.TODO(), existingUserId, existingWithdrawal1.Sum+existingWithdrawal2.Sum, nil)
			moneyService.Withdraw(context.TODO(), existingUserId, &existingWithdrawal1)
			moneyService.Withdraw(context.TODO(), existingUserId, &existingWithdrawal2)

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", tc.inputUserId.String()).
				Execute(http.MethodGet, srv.URL+"/api/user/withdrawals")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			if tc.expectedWithdrawalsNum != 0 {
				var respWithdrawals []withdrawal.Withdrawal
				json.Unmarshal(resp.Body(), &respWithdrawals)
				assert.Equal(t, tc.expectedWithdrawalsNum, len(respWithdrawals), "withdrawals not equal")
			}
		})
	}
}

type InputWithdrawal struct {
	OrderId string  `json:"order"`
	Sum     float64 `json:"sum"`
}

func TestPostWithdraw(t *testing.T) {
	existingUserId := uuid.New()
	existingBalanceCurrent := float64(200)
	existingWithdrawal := withdrawal.Withdrawal{
		Id:      uuid.New(),
		OrderId: "1115",
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
				OrderId: "1321",
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
				OrderId: "1321",
				Sum:     0,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusBadRequest,
		},
		{
			testName:              "not enough money on balance",
			inputIdempotencyToken: uuid.NewString(),
			inputWithdrawal: &InputWithdrawal{
				OrderId: "1321",
				Sum:     existingBalanceCurrent + 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusPaymentRequired,
		},
		{
			testName:              "invalid order id format",
			inputIdempotencyToken: uuid.NewString(),
			inputWithdrawal: &InputWithdrawal{
				OrderId: "1322",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusUnprocessableEntity,
		},
		{
			testName:              "existing idempotency token",
			inputIdempotencyToken: existingWithdrawal.Id.String(),
			inputWithdrawal: &InputWithdrawal{
				OrderId: "1321",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusOK,
		},
		{
			testName:              "invalid idempotency token",
			inputIdempotencyToken: "invalid",
			inputWithdrawal: &InputWithdrawal{
				OrderId: "1321",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 1,
			expectedCode:           http.StatusBadRequest,
		},
		{
			testName:              "empty idempotency token",
			inputIdempotencyToken: "",
			inputWithdrawal: &InputWithdrawal{
				OrderId: "1321",
				Sum:     existingBalanceCurrent - 100,
			},
			expectedWithdrawalsNum: 2,
			expectedCode:           http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := memstorage.NewMemStorage()
			moneyService := moneyservice.NewMoneyService(storage)
			handlers := NewMoneyHandlers(moneyService)
			router := mockRouter(handlers)
			srv := httptest.NewServer(router)
			defer srv.Close()
			client := resty.New()

			moneyService.AddAccrual(context.TODO(), existingUserId, existingBalanceCurrent+existingWithdrawal.Sum, nil)
			moneyService.Withdraw(context.TODO(), existingUserId, &existingWithdrawal)

			var req []byte
			if tc.inputWithdrawal != nil {
				req, _ = json.Marshal(*tc.inputWithdrawal)
			} else {
				req, _ = json.Marshal(map[string]string{})
			}

			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-User-Id", existingUserId.String()).
				SetHeader("Idempotency-Key", tc.inputIdempotencyToken).
				SetBody(req).
				Execute(http.MethodPost, srv.URL+"/api/user/balance/withdraw")
			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			userWithdrawals, _ := moneyService.GetWithdrawals(context.TODO(), existingUserId)
			assert.Equal(t, tc.expectedWithdrawalsNum, len(userWithdrawals), "num of withdrawals don't match")
		})
	}
}
