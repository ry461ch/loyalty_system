package memstorage

import (
	"context"

	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/users"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type MemStorage struct {
	*withdrawalmemstorage.WithdrawalMemStorage
	*ordermemstorage.OrderMemStorage
	*usermemstorage.UserMemStorage
	*balancememstorage.BalanceMemStorage
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		WithdrawalMemStorage: withdrawalmemstorage.NewWithdrawalMemStorage(),
		OrderMemStorage:      ordermemstorage.NewOrderMemStorage(),
		UserMemStorage:       usermemstorage.NewUserMemStorage(),
		BalanceMemStorage:    balancememstorage.NewBalanceMemStorage(),
	}
}

func (ms *MemStorage) Initialize(ctx context.Context) error {
	ms.InitializeOrderMemStorage(ctx)
	ms.InitializeUserMemStorage(ctx)
	ms.InitializeWithdrawalMemStorage(ctx)
	ms.InitializeBalanceMemStorage(ctx)
	return nil
}

func (ms *MemStorage) BeginTx(ctx context.Context) (*transaction.Trx, error) {
	return transaction.BeginTx(ctx, nil)
}
