package withdrawalmemstorage

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type WithdrawalMemStorage struct {
	usersToWithdrawalsMap sync.Map // map[uuid.UUID]map[uuid.UUID]withdrawal.Withdrawal
	withdrawalsToUsersMap sync.Map // map[uuid.UUID]uuid.UUID
}

func NewWithdrawalMemStorage() *WithdrawalMemStorage {
	return &WithdrawalMemStorage{}
}

func (wms *WithdrawalMemStorage) InsertWithdrawal(ctx context.Context, inputWithdrawal *withdrawal.Withdrawal, trx *transaction.Trx) error {
	val, ok := wms.usersToWithdrawalsMap.Load(*inputWithdrawal.UserID)
	if !ok {
		val = map[uuid.UUID]withdrawal.Withdrawal{}
	}
	userWithdrawals := val.(map[uuid.UUID]withdrawal.Withdrawal)

	createdAt := time.Now().UTC()
	userWithdrawals[*inputWithdrawal.ID] = withdrawal.Withdrawal{
		CreatedAt: &createdAt,
		ID:        inputWithdrawal.ID,
		OrderID:   inputWithdrawal.OrderID,
		Sum:       inputWithdrawal.Sum,
		UserID:    inputWithdrawal.UserID,
	}
	wms.usersToWithdrawalsMap.Store(*inputWithdrawal.UserID, userWithdrawals)
	wms.withdrawalsToUsersMap.Store(*inputWithdrawal.ID, *inputWithdrawal.UserID)
	return nil
}

func (wms *WithdrawalMemStorage) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]withdrawal.Withdrawal, error) {
	val, ok := wms.usersToWithdrawalsMap.Load(userID)
	if !ok {
		val = map[uuid.UUID]withdrawal.Withdrawal{}
	}
	userWithdrawals := val.(map[uuid.UUID]withdrawal.Withdrawal)
	resultWithdrawals := []withdrawal.Withdrawal{}
	for _, userWithdrawal := range userWithdrawals {
		resultWithdrawals = append(resultWithdrawals, userWithdrawal)
	}

	slices.SortFunc(resultWithdrawals, func(left, right withdrawal.Withdrawal) int {
		return right.CreatedAt.Compare(*left.CreatedAt)
	})
	return resultWithdrawals, nil
}

func (wms *WithdrawalMemStorage) GetWithdrawal(ctx context.Context, ID uuid.UUID) (*withdrawal.Withdrawal, error) {
	userVal, ok := wms.withdrawalsToUsersMap.Load(ID)
	if !ok {
		return nil, exceptions.ErrWithdrawalNotFound
	}
	userID := userVal.(uuid.UUID)

	withdrawalsVal, ok := wms.usersToWithdrawalsMap.Load(userID)
	if !ok {
		withdrawalsVal = map[uuid.UUID]withdrawal.Withdrawal{}
	}
	userWithdrawals := withdrawalsVal.(map[uuid.UUID]withdrawal.Withdrawal)
	withdrawalInDB := userWithdrawals[ID]
	return &withdrawalInDB, nil
}

func (*WithdrawalMemStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, nil)
}
