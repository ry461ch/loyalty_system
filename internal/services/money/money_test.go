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
	existingUserId := uuid.New()
	createdAt1, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	existingWithdrawalId1 := uuid.New()
	existingWithdrawalId2 := uuid.New()
	createdAt2, _ := time.Parse(time.RFC3339, "2020-12-10T16:09:53Z")
	existingWithdrawals := []withdrawal.Withdrawal{
		{
			Id:        &existingWithdrawalId1,
			OrderId:   "1115",
			UserId:    &existingUserId,
			Sum:       500,
			CreatedAt: &createdAt1,
		},
		{
			Id:        &existingWithdrawalId2,
			OrderId:   "1313",
			UserId:    &existingUserId,
			Sum:       400,
			CreatedAt: &createdAt2,
		},
	}

	testCases := []struct {
		testName            string
		userId              uuid.UUID
		expectedWithdrawals []withdrawal.Withdrawal
	}{
		{
			testName:            "existing user",
			userId:              existingUserId,
			expectedWithdrawals: existingWithdrawals,
		},
		{
			testName:            "new user",
			userId:              uuid.New(),
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
			userWithdrawals, _ := service.GetWithdrawals(context.TODO(), tc.userId)
			assert.Equal(t, len(tc.expectedWithdrawals), len(userWithdrawals), "num of withdrawals don't match")
			for idx, existingWithdrawal := range tc.expectedWithdrawals {
				assert.Equal(t, tc.expectedWithdrawals[idx].Id, existingWithdrawal.Id, "withdrawal id don't match")
				assert.Equal(t, tc.expectedWithdrawals[idx].Sum, existingWithdrawal.Sum, "withdrawal sum don't match")
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	existingUserId := uuid.New()
	existingBalance := balance.Balance{
		Current:   500,
		Withdrawn: 300,
	}

	testCases := []struct {
		testName        string
		userId          uuid.UUID
		expectedBalance balance.Balance
	}{
		{
			testName: "existing user",
			userId:   existingUserId,
			expectedBalance: balance.Balance{
				Current:   existingBalance.Current,
				Withdrawn: existingBalance.Withdrawn,
			},
		},
		{
			testName:        "new user",
			userId:          uuid.New(),
			expectedBalance: balance.Balance{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			balanceStorage.AddBalance(context.TODO(), existingUserId, existingBalance.Current+existingBalance.Withdrawn, nil)
			balanceStorage.ReduceBalance(context.TODO(), existingUserId, existingBalance.Withdrawn, nil)

			service := NewMoneyService(balanceStorage, withdrawalStorage)
			userBalance, _ := service.GetBalance(context.TODO(), tc.userId)
			assert.Equal(t, tc.expectedBalance, *userBalance, "balances don't match")
		})
	}
}

func TestAddAccrual(t *testing.T) {
	existingUserId := uuid.New()
	existingBalance := balance.Balance{
		Current:   500,
		Withdrawn: 300,
	}

	testCases := []struct {
		testName        string
		userId          uuid.UUID
		accrual         float64
		expectedBalance *balance.Balance
	}{
		{
			testName: "existing user",
			userId:   existingUserId,
			accrual:  200,
			expectedBalance: &balance.Balance{
				Current:   existingBalance.Current + 200,
				Withdrawn: existingBalance.Withdrawn,
			},
		},
		{
			testName: "new user",
			userId:   uuid.New(),
			accrual:  200,
			expectedBalance: &balance.Balance{
				Current:   200,
				Withdrawn: 0,
			},
		},
		{
			testName:        "bad amount",
			userId:          uuid.New(),
			accrual:         0,
			expectedBalance: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			balanceStorage.AddBalance(context.TODO(), existingUserId, existingBalance.Current+existingBalance.Withdrawn, nil)
			balanceStorage.ReduceBalance(context.TODO(), existingUserId, existingBalance.Withdrawn, nil)

			service := NewMoneyService(balanceStorage, withdrawalStorage)
			err := service.AddAccrual(context.TODO(), tc.userId, tc.accrual, nil)
			if tc.expectedBalance != nil {
				balanceInDB, _ := balanceStorage.GetBalance(context.TODO(), tc.userId)
				assert.Equal(t, *tc.expectedBalance, *balanceInDB, "balances don't match")
			} else {
				assert.ErrorIs(t, err, exceptions.NewBalanceBadAmountFormatError(), "exceptions don't match")
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	existingUserId := uuid.New()
	existingBalance := balance.Balance{
		Current:   300,
		Withdrawn: 300,
	}
	existingWithdrawalID := uuid.New()
	createdAt := time.Now()
	existingWithdrawal := withdrawal.Withdrawal{
		Id:        &existingWithdrawalID,
		UserId:    &existingUserId,
		OrderId:   "1321",
		Sum:       300,
		CreatedAt: &createdAt,
	}
	newWithdrawalId := uuid.New()

	testCases := []struct {
		testName        string
		inputWithdrawal withdrawal.Withdrawal
		expectedError   error
		expectedBalance balance.Balance
	}{
		{
			testName: "successfully withdrawn",
			inputWithdrawal: withdrawal.Withdrawal{
				Id:      &newWithdrawalId,
				OrderId: "1115",
				UserId:  &existingUserId,
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
				Id:      existingWithdrawal.Id,
				OrderId: "1321",
				UserId:  &existingUserId,
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
				Id:      &newWithdrawalId,
				OrderId: "1115",
				UserId:  &existingUserId,
				Sum:     existingBalance.Current + 100,
			},
			expectedError:   exceptions.NewBalanceNotEnoughBalanceError(),
			expectedBalance: existingBalance,
		},
		{
			testName: "bad amount format",
			inputWithdrawal: withdrawal.Withdrawal{
				Id:      &newWithdrawalId,
				OrderId: "1115",
				UserId:  &existingUserId,
				Sum:     0,
			},
			expectedError:   exceptions.NewBalanceBadAmountFormatError(),
			expectedBalance: existingBalance,
		},
		{
			testName: "bad order id format",
			inputWithdrawal: withdrawal.Withdrawal{
				Id:      &newWithdrawalId,
				OrderId: "1114",
				UserId:  &existingUserId,
				Sum:     100,
			},
			expectedError:   exceptions.NewOrderBadIdFormatError(),
			expectedBalance: existingBalance,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			balanceStorage := balancememstorage.NewBalanceMemStorage()
			withdrawalStorage := withdrawalmemstorage.NewWithdrawalMemStorage()
			balanceStorage.AddBalance(context.TODO(), existingUserId, existingBalance.Current+existingBalance.Withdrawn, nil)
			balanceStorage.ReduceBalance(context.TODO(), existingUserId, existingBalance.Withdrawn, nil)
			withdrawalStorage.InsertWithdrawal(context.TODO(), &existingWithdrawal, nil)

			service := NewMoneyService(balanceStorage, withdrawalStorage)
			err := service.Withdraw(context.TODO(), &tc.inputWithdrawal)
			if tc.expectedError == nil {
				assert.Nil(t, err, "error was unexpected")
				balanceInDB, _ := balanceStorage.GetBalance(context.TODO(), existingUserId)
				assert.Equal(t, tc.expectedBalance, *balanceInDB, "balances don't match")
			} else {
				assert.ErrorIs(t, err, tc.expectedError, "exceptions don't match")
			}
		})
	}
}
