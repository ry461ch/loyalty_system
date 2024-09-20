package balancememstorage

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type BalanceMemStorage struct {
	balances sync.Map // map[string]balance.Balance
}

func NewBalanceMemStorage() *BalanceMemStorage {
	return &BalanceMemStorage{}
}

func (bms *BalanceMemStorage) InitializeBalanceMemStorage(ctx context.Context) error {
	return nil
}

func (bms *BalanceMemStorage) GetBalance(ctx context.Context, userID uuid.UUID) (*balance.Balance, error) {
	val, ok := bms.balances.Load(userID)
	if !ok {
		val = balance.Balance{}
	}
	userBalance := val.(balance.Balance)
	return &userBalance, nil
}

func (bms *BalanceMemStorage) ReduceBalance(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error {
	val, ok := bms.balances.Load(userID)
	if !ok {
		val = balance.Balance{}
	}
	userBalance := val.(balance.Balance)
	userBalance.Current -= amount
	userBalance.Withdrawn += amount
	bms.balances.Store(userID, userBalance)
	return nil
}

func (bms *BalanceMemStorage) AddBalance(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error {
	val, ok := bms.balances.Load(userID)
	if !ok {
		val = balance.Balance{}
	}
	userBalance := val.(balance.Balance)
	userBalance.Current += amount
	bms.balances.Store(userID, userBalance)
	return nil
}

func (*BalanceMemStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, nil)
}
