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

func (wms *WithdrawalMemStorage) InitializeWithdrawalMemStorage(ctx context.Context) error {
	return nil
}

func (wms *WithdrawalMemStorage) InsertWithdrawal(ctx context.Context, inputWithdrawal *withdrawal.Withdrawal, trx *transaction.Trx) error {
	val, ok := wms.usersToWithdrawalsMap.Load(*inputWithdrawal.UserId)
	if !ok {
		val = map[uuid.UUID]withdrawal.Withdrawal{}
	}
	userWithdrawals := val.(map[uuid.UUID]withdrawal.Withdrawal)

	createdAt := time.Now()
	userWithdrawals[*inputWithdrawal.Id] = withdrawal.Withdrawal{
		CreatedAt: &createdAt,
		Id:        inputWithdrawal.Id,
		OrderId:   inputWithdrawal.OrderId,
		Sum:       inputWithdrawal.Sum,
		UserId:    inputWithdrawal.UserId,
	}
	wms.usersToWithdrawalsMap.Store(*inputWithdrawal.UserId, userWithdrawals)
	wms.withdrawalsToUsersMap.Store(*inputWithdrawal.Id, *inputWithdrawal.UserId)
	return nil
}

func (wms *WithdrawalMemStorage) GetWithdrawals(ctx context.Context, userId uuid.UUID) ([]withdrawal.Withdrawal, error) {
	val, ok := wms.usersToWithdrawalsMap.Load(userId)
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

func (wms *WithdrawalMemStorage) GetWithdrawal(ctx context.Context, Id uuid.UUID) (*withdrawal.Withdrawal, error) {
	userVal, ok := wms.withdrawalsToUsersMap.Load(Id)
	if !ok {
		return nil, exceptions.NewWithdrawalNotFoundError()
	}
	userId := userVal.(uuid.UUID)

	withdrawalsVal, ok := wms.usersToWithdrawalsMap.Load(userId)
	if !ok {
		withdrawalsVal = map[uuid.UUID]withdrawal.Withdrawal{}
	}
	userWithdrawals := withdrawalsVal.(map[uuid.UUID]withdrawal.Withdrawal)
	withdrawalInDb := userWithdrawals[Id]
	return &withdrawalInDb, nil
}

func (*WithdrawalMemStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, nil)
}
