package withdrawalmemstorage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
)

func TestInsertWithdrawal(t *testing.T) {
	existingUserID := uuid.New()
	existingWithdrawalID := uuid.New()
	createdAt, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	existingWithdrawal := withdrawal.Withdrawal{
		ID:        &existingWithdrawalID,
		OrderID:   "1115",
		Sum:       500,
		CreatedAt: &createdAt,
	}
	newWithdrawalID := uuid.New()
	newUserID := uuid.New()

	testCases := []struct {
		testName               string
		inputWithdrawal        withdrawal.Withdrawal
		expectedWithdrawalsNum int
	}{
		{
			testName: "existing userId",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      &newWithdrawalID,
				UserID:  &existingUserID,
				OrderID: "1313",
				Sum:     400,
			},
			expectedWithdrawalsNum: 2,
		},
		{
			testName: "new userId",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      &newWithdrawalID,
				UserID:  &newUserID,
				OrderID: "1313",
				Sum:     400,
			},
			expectedWithdrawalsNum: 1,
		},
		{
			testName: "existing withdrawal",
			inputWithdrawal: withdrawal.Withdrawal{
				ID:      &existingWithdrawalID,
				UserID:  &existingUserID,
				OrderID: "1313",
				Sum:     400,
			},
			expectedWithdrawalsNum: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewWithdrawalMemStorage()
			storage.usersToWithdrawalsMap.Store(existingUserID, map[uuid.UUID]withdrawal.Withdrawal{existingWithdrawalID: existingWithdrawal})
			storage.withdrawalsToUsersMap.Store(*existingWithdrawal.ID, existingUserID)

			storage.InsertWithdrawal(context.TODO(), &tc.inputWithdrawal, nil)
			userWithdrawals, _ := storage.usersToWithdrawalsMap.Load(*tc.inputWithdrawal.UserID)
			assert.Equal(t, tc.expectedWithdrawalsNum, len(userWithdrawals.(map[uuid.UUID]withdrawal.Withdrawal)), "num of withdrawals doesn't match")
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	existingUserID := uuid.New()
	createdAt1, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	createdAt2, _ := time.Parse(time.RFC3339, "2020-12-10T16:09:53Z")
	existingWithdrawalID1 := uuid.New()
	existingWithdrawalID2 := uuid.New()
	existingWithdrawal1 := withdrawal.Withdrawal{
		ID:        &existingWithdrawalID1,
		OrderID:   "1115",
		Sum:       500,
		CreatedAt: &createdAt1,
	}
	existingWithdrawal2 := withdrawal.Withdrawal{
		ID:        &existingWithdrawalID2,
		OrderID:   "1313",
		Sum:       400,
		CreatedAt: &createdAt2,
	}
	existingWithdrawals := map[uuid.UUID]withdrawal.Withdrawal{
		existingWithdrawalID1: existingWithdrawal1,
		existingWithdrawalID2: existingWithdrawal2,
	}

	testCases := []struct {
		testName            string
		userID              uuid.UUID
		expectedWithdrawals []withdrawal.Withdrawal
	}{
		{
			testName: "existing user id",
			userID:   existingUserID,
			expectedWithdrawals: []withdrawal.Withdrawal{
				existingWithdrawal2,
				existingWithdrawal1,
			},
		},
		{
			testName:            "new user id",
			userID:              uuid.New(),
			expectedWithdrawals: []withdrawal.Withdrawal{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewWithdrawalMemStorage()
			storage.usersToWithdrawalsMap.Store(existingUserID, existingWithdrawals)
			storage.withdrawalsToUsersMap.Store(*existingWithdrawal1.ID, existingUserID)
			storage.withdrawalsToUsersMap.Store(*existingWithdrawal2.ID, existingUserID)

			userWithdrawals, _ := storage.GetWithdrawals(context.TODO(), tc.userID)
			assert.Equal(t, tc.expectedWithdrawals, userWithdrawals, "withdrawals don't match")
		})
	}
}

func TestGetWithdrawal(t *testing.T) {
	existingUserID := uuid.New()
	createdAt, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	existingWithdrawalID := uuid.New()
	existingWithdrawal := withdrawal.Withdrawal{
		ID:        &existingWithdrawalID,
		OrderID:   "1115",
		Sum:       500,
		CreatedAt: &createdAt,
	}

	testCases := []struct {
		testName     string
		withdrawalID uuid.UUID
		expectedErr  error
	}{
		{
			testName:     "existing withdrawalId",
			withdrawalID: existingWithdrawalID,
			expectedErr:  nil,
		},
		{
			testName:     "new withdrawalId",
			withdrawalID: uuid.New(),
			expectedErr:  exceptions.ErrWithdrawalNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewWithdrawalMemStorage()
			storage.usersToWithdrawalsMap.Store(existingUserID, map[uuid.UUID]withdrawal.Withdrawal{*existingWithdrawal.ID: existingWithdrawal})
			storage.withdrawalsToUsersMap.Store(*existingWithdrawal.ID, existingUserID)

			userWithdrawal, err := storage.GetWithdrawal(context.TODO(), tc.withdrawalID)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, exceptions.ErrWithdrawalNotFound, "errors don't match")
			} else {
				assert.Equal(t, existingWithdrawal, *userWithdrawal, "withdrawals don't match")
			}
		})
	}
}
