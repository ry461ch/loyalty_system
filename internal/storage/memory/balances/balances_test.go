package balancememstorage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
)

func TestReduceBalance(t *testing.T) {
	existingUserId := uuid.New()
	existingBalance := balance.Balance{
		Current:   500,
		Withdrawn: 300,
	}

	testCases := []struct {
		testName        string
		userId          uuid.UUID
		withdrawal      float64
		expectedBalance balance.Balance
	}{
		{
			testName:   "existing user",
			userId:     existingUserId,
			withdrawal: 200,
			expectedBalance: balance.Balance{
				Current:   existingBalance.Current - 200,
				Withdrawn: existingBalance.Withdrawn + 200,
			},
		},
		{
			testName:   "new user",
			userId:     uuid.New(),
			withdrawal: 200,
			expectedBalance: balance.Balance{
				Current:   -200,
				Withdrawn: 200,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewBalanceMemStorage()
			storage.balances.Store(existingUserId, existingBalance)

			storage.ReduceBalance(context.TODO(), tc.userId, tc.withdrawal, nil)
			resultBalance, _ := storage.balances.Load(tc.userId)
			assert.Equal(t, tc.expectedBalance, resultBalance, "balances don't match")
		})
	}
}

func TestAddBalance(t *testing.T) {
	existingUserId := uuid.New()
	existingBalance := balance.Balance{
		Current:   500,
		Withdrawn: 300,
	}

	testCases := []struct {
		testName        string
		userId          uuid.UUID
		withdrawal      float64
		expectedBalance balance.Balance
	}{
		{
			testName:   "existing user",
			userId:     existingUserId,
			withdrawal: 200,
			expectedBalance: balance.Balance{
				Current:   existingBalance.Current + 200,
				Withdrawn: existingBalance.Withdrawn,
			},
		},
		{
			testName:   "new user",
			userId:     uuid.New(),
			withdrawal: 200,
			expectedBalance: balance.Balance{
				Current: 200,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewBalanceMemStorage()
			storage.balances.Store(existingUserId, existingBalance)

			storage.AddBalance(context.TODO(), tc.userId, tc.withdrawal, nil)
			resultBalance, _ := storage.balances.Load(tc.userId)
			assert.Equal(t, tc.expectedBalance, resultBalance, "balances don't match")
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
			storage := NewBalanceMemStorage()
			storage.balances.Store(existingUserId, existingBalance)

			resultBalance, _ := storage.GetBalance(context.TODO(), tc.userId)
			assert.Equal(t, tc.expectedBalance, *resultBalance, "balances don't match")
		})
	}
}
