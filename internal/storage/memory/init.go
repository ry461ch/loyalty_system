package memstorage

import (
	"github.com/ry461ch/loyalty_system/internal/storage/memory/balances"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/orders"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/users"
	"github.com/ry461ch/loyalty_system/internal/storage/memory/withdrawals"
)

type MemStorage struct {
	OrderStorage      *ordermemstorage.OrderMemStorage
	WithdrawalStorage *withdrawalmemstorage.WithdrawalMemStorage
	BalanceStorage    *balancememstorage.BalanceMemStorage
	UserStorage       *usermemstorage.UserMemStorage
}

func NewPGStorage() *MemStorage {
	return &MemStorage{
		OrderStorage:      ordermemstorage.NewOrderMemStorage(),
		UserStorage:       usermemstorage.NewUserMemStorage(),
		BalanceStorage:    balancememstorage.NewBalanceMemStorage(),
		WithdrawalStorage: withdrawalmemstorage.NewWithdrawalMemStorage(),
	}
}
