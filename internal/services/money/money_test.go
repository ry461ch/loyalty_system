package moneyservice

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

func TestGetWithdrawals(t *testing.T) {
	existingUserID := uuid.New()
	createdAt1, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	existingWithdrawalID1 := uuid.New()
	existingWithdrawalID2 := uuid.New()
	createdAt2, _ := time.Parse(time.RFC3339, "2020-12-10T16:09:53Z")
	existingWithdrawals := []withdrawal.Withdrawal{
		{
			ID:        &existingWithdrawalID1,
			OrderID:   "1115",
			UserID:    &existingUserID,
			Sum:       500,
			CreatedAt: &createdAt1,
		},
		{
			ID:        &existingWithdrawalID2,
			OrderID:   "1313",
			UserID:    &existingUserID,
			Sum:       400,
			CreatedAt: &createdAt2,
		},
	}

	testCases := []struct {
		testName            string
		userID              uuid.UUID
		expectedWithdrawals []withdrawal.Withdrawal
	}{
		{
			testName:            "existing user",
			userID:              existingUserID,
			expectedWithdrawals: existingWithdrawals,
		},
		{
			testName:            "new user",
			userID:              uuid.New(),
			expectedWithdrawals: []withdrawal.Withdrawal{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			for _, existingWithdrawal := range existingWithdrawals {
				withdrawalStorage.InsertWithdrawal(context.TODO(), &existingWithdrawal, nil)
			}

			service := NewMoneyService(balanceStorage, withdrawalStorage)
			userWithdrawals, _ := service.GetWithdrawals(context.TODO(), tc.userID)
			assert.Equal(t, len(tc.expectedWithdrawals), len(userWithdrawals), "num of withdrawals don't match")
			for idx, existingWithdrawal := range tc.expectedWithdrawals {
				assert.Equal(t, tc.expectedWithdrawals[idx].ID, existingWithdrawal.ID, "withdrawal id don't match")
				assert.Equal(t, tc.expectedWithdrawals[idx].Sum, existingWithdrawal.Sum, "withdrawal sum don't match")
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	existingUserID := uuid.New()
	existingBalance := balance.Balance{
		Current:   500,
		Withdrawn: 300,
	}

	testCases := []struct {
		testName        string
		userID          uuid.UUID
		expectedBalance balance.Balance
	}{
		{
			testName: "existing user",
			userID:   existingUserID,
			expectedBalance: balance.Balance{
				Current:   existingBalance.Current,
				Withdrawn: existingBalance.Withdrawn,
			},
		},
		{
			testName:        "new user",
			userID:          uuid.New(),
			expectedBalance: balance.Balance{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			balanceStorage.AddBalance(context.TODO(), existingUserID, existingBalance.Current+existingBalance.Withdrawn, nil)
			balanceStorage.ReduceBalance(context.TODO(), existingUserID, existingBalance.Withdrawn, nil)

			service := NewMoneyService(balanceStorage, withdrawalStorage)
			userBalance, _ := service.GetBalance(context.TODO(), tc.userID)
			assert.Equal(t, tc.expectedBalance, *userBalance, "balances don't match")
		})
	}
}

func TestWithdraw(t *testing.T) {
	existingUserID := uuid.New()
	existingBalance := balance.Balance{
		Current:   300,
		Withdrawn: 300,
	}
	existingWithdrawalID := uuid.New()
	createdAt := time.Now()
	existingWithdrawal := withdrawal.Withdrawal{
		ID:        &existingWithdrawalID,
		UserID:    &existingUserID,
		OrderID:   "1321",
		Sum:       300,
		CreatedAt: &createdAt,
	}
	newWithdrawalID := uuid.New()

	testCases := []struct {
		testName        string
		inputWithdrawal withdrawal.Withdrawal
		expectedError   error
		expectedBalance balance.Balance
	}{
		{
			testName: "successfully withdrawn",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      &newWithdrawalID,
				OrderID: "1115",
				UserID:  &existingUserID,
				Sum:     200,
			},
			expectedError: nil,
			expectedBalance: balance.Balance{
				Current:   existingBalance.Current - 200,
				Withdrawn: existingBalance.Withdrawn + 200,
			},
		},
		{
			testName: "existing withdrawal",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      existingWithdrawal.ID,
				OrderID: "1321",
				UserID:  &existingUserID,
				Sum:     300,
			},
			expectedError: nil,
			expectedBalance: balance.Balance{
				Current:   existingBalance.Current,
				Withdrawn: existingBalance.Withdrawn,
			},
		},
		{
			testName: "not enough balance",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      &newWithdrawalID,
				OrderID: "1115",
				UserID:  &existingUserID,
				Sum:     existingBalance.Current + 100,
			},
			expectedError:   exceptions.NewBalanceNotEnoughBalanceError(),
			expectedBalance: existingBalance,
		},
		{
			testName: "bad amount format",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      &newWithdrawalID,
				OrderID: "1115",
				UserID:  &existingUserID,
				Sum:     0,
			},
			expectedError:   exceptions.NewBalanceBadAmountFormatError(),
			expectedBalance: existingBalance,
		},
		{
			testName: "bad order id format",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      &newWithdrawalID,
				OrderID: "1114",
				UserID:  &existingUserID,
				Sum:     100,
			},
			expectedError:   exceptions.NewOrderBadIDFormatError(),
			expectedBalance: existingBalance,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			balanceStorage.AddBalance(context.TODO(), existingUserID, existingBalance.Current+existingBalance.Withdrawn, nil)
			balanceStorage.ReduceBalance(context.TODO(), existingUserID, existingBalance.Withdrawn, nil)
			withdrawalStorage.InsertWithdrawal(context.TODO(), &existingWithdrawal, nil)

			service := NewMoneyService(balanceStorage, withdrawalStorage)
			err := service.Withdraw(context.TODO(), &tc.inputWithdrawal)
			if tc.expectedError == nil {
				assert.Nil(t, err, "error was unexpected")
				balanceInDB, _ := balanceStorage.GetBalance(context.TODO(), existingUserID)
				assert.Equal(t, tc.expectedBalance, *balanceInDB, "balances don't match")
			} else {
				assert.ErrorIs(t, err, tc.expectedError, "exceptions don't match")
			}
		})
	}
}

func TestAddAccrual(t *testing.T) {
	existingUserID := uuid.New()
	existingBalance := balance.Balance{
		Current:   500,
		Withdrawn: 300,
	}

	testCases := []struct {
		testName        string
		userID          uuid.UUID
		accrual         float64
		expectedBalance *balance.Balance
	}{
		{
			testName: "existing user",
			userID:   existingUserID,
			accrual:  200,
			expectedBalance: &balance.Balance{
				Current:   existingBalance.Current + 200,
				Withdrawn: existingBalance.Withdrawn,
			},
		},
		{
			testName: "new user",
			userID:   uuid.New(),
			accrual:  200,
			expectedBalance: &balance.Balance{
				Current:   200,
				Withdrawn: 0,
			},
		},
		{
			testName:        "bad amount",
			userID:          uuid.New(),
			accrual:         0,
			expectedBalance: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			balanceStorage.AddBalance(context.TODO(), existingUserID, existingBalance.Current+existingBalance.Withdrawn, nil)
			balanceStorage.ReduceBalance(context.TODO(), existingUserID, existingBalance.Withdrawn, nil)

			service := NewMoneyService(balanceStorage, withdrawalStorage)
			err := service.AddAccrual(context.TODO(), tc.userID, tc.accrual, nil)
			if tc.expectedBalance != nil {
				balanceInDB, _ := balanceStorage.GetBalance(context.TODO(), tc.userID)
				assert.Equal(t, *tc.expectedBalance, *balanceInDB, "balances don't match")
			} else {
				assert.ErrorIs(t, err, exceptions.NewBalanceBadAmountFormatError(), "exceptions don't match")
			}
		})
	}
}
