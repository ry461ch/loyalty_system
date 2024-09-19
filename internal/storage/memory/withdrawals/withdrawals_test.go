package withdrawalmemstorage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
)

func TestInsertWithdrawal(t *testing.T) {
	existingUserId := uuid.New()
	createdAt, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	existingWithdrawal := withdrawal.Withdrawal{
		Id: 	   uuid.New(),
		OrderId:   "1115",
		Sum:       500,
		CreatedAt: createdAt,
	}

	testCases := []struct {
		testName               string
		userId                 uuid.UUID
		inputWithdrawal        withdrawal.Withdrawal
		expectedWithdrawalsNum int
	}{
		{
			testName: "existing userId",
			userId:   existingUserId,
			inputWithdrawal: withdrawal.Withdrawal{
				Id: 		uuid.New(),
				OrderId:  "1313",
				Sum: 400,
			},
			expectedWithdrawalsNum: 2,
		},
		{
			testName: "new userId",
			userId:   uuid.New(),
			inputWithdrawal: withdrawal.Withdrawal{
				Id: 		uuid.New(),
				OrderId:  "1313",
				Sum: 400,
			},
			expectedWithdrawalsNum: 1,
		},
		{
			testName: "existing withdrawal",
			userId:   existingUserId,
			inputWithdrawal: withdrawal.Withdrawal{
				Id: 		existingWithdrawal.Id,
				OrderId:  "1313",
				Sum: 400,
			},
			expectedWithdrawalsNum: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewWithdrawalMemStorage()
			storage.usersToWithdrawalsMap.Store(existingUserId, map[uuid.UUID]withdrawal.Withdrawal{existingWithdrawal.Id: existingWithdrawal})
			storage.withdrawalsToUsersMap.Store(existingWithdrawal.Id, existingUserId)

			storage.InsertWithdrawl(context.TODO(), tc.userId, &tc.inputWithdrawal, nil)
			userWithdrawals, _ := storage.usersToWithdrawalsMap.Load(tc.userId)
			assert.Equal(t, tc.expectedWithdrawalsNum, len(userWithdrawals.(map[uuid.UUID]withdrawal.Withdrawal)), "num of withdrawals doesn't match")
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	existingUserId := uuid.New()
	createdAt1, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	createdAt2, _ := time.Parse(time.RFC3339, "2020-12-10T16:09:53Z")
	existingWithdrawal1 := withdrawal.Withdrawal{
		Id: 	   uuid.New(),
		OrderId:   "1115",
		Sum:       500,
		CreatedAt: createdAt1,
	}
	existingWithdrawal2 := withdrawal.Withdrawal{
		Id: 	   uuid.New(),
		OrderId:   "1313",
		Sum:       400,
		CreatedAt: createdAt2,
	}
	existingWithdrawals := map[uuid.UUID]withdrawal.Withdrawal{
		existingWithdrawal1.Id: existingWithdrawal1,
		existingWithdrawal2.Id: existingWithdrawal2,
	}

	testCases := []struct {
		testName            string
		userId              uuid.UUID
		expectedWithdrawals []withdrawal.Withdrawal
	}{
		{
			testName:            "existing userId",
			userId:              existingUserId,
			expectedWithdrawals: []withdrawal.Withdrawal{
				existingWithdrawal2,
				existingWithdrawal1,
			},
		},
		{
			testName:            "new userId",
			userId:              uuid.New(),
			expectedWithdrawals: []withdrawal.Withdrawal{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewWithdrawalMemStorage()
			storage.usersToWithdrawalsMap.Store(existingUserId, existingWithdrawals)
			storage.withdrawalsToUsersMap.Store(existingWithdrawal1.Id, existingUserId)
			storage.withdrawalsToUsersMap.Store(existingWithdrawal2.Id, existingUserId)

			userWithdrawals, _ := storage.GetWithdrawals(context.TODO(), tc.userId)
			assert.Equal(t, tc.expectedWithdrawals, userWithdrawals, "withdrawals don't match")
		})
	}
}

func TestGetWithdrawal(t *testing.T) {
	existingUserId := uuid.New()
	createdAt, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53Z")
	existingWithdrawal := withdrawal.Withdrawal{
		Id: 	   uuid.New(),
		OrderId:   "1115",
		Sum:       500,
		CreatedAt: createdAt,
	}

	testCases := []struct {
		testName            string
		withdrawalId              uuid.UUID
		expectedErr		    error
	}{
		{
			testName:            "existing withdrawalId",
			withdrawalId:        existingWithdrawal.Id,
			expectedErr: 		 nil,
		},
		{
			testName:            "new withdrawalId",
			withdrawalId:        uuid.New(),
			expectedErr:  		 exceptions.NewWithdrawalNotFoundError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			storage := NewWithdrawalMemStorage()
			storage.usersToWithdrawalsMap.Store(existingUserId, map[uuid.UUID]withdrawal.Withdrawal{existingWithdrawal.Id: existingWithdrawal})
			storage.withdrawalsToUsersMap.Store(existingWithdrawal.Id, existingUserId)

			userWithdrawal, err := storage.GetWithdrawal(context.TODO(), tc.withdrawalId)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, exceptions.NewWithdrawalNotFoundError(), "errors don't match")
			} else {
				assert.Equal(t, existingWithdrawal, *userWithdrawal, "withdrawals don't match")
			}
		})
	}
}
