package moneyservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/helpers/order"
	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
	"github.com/ry461ch/loyalty_system/internal/storage"
	"github.com/ry461ch/loyalty_system/internal/storage/transaction"
)

type MoneyService struct {
	moneyStorage storage.MoneyStorage
}

func NewMoneyService(moneyStorage storage.MoneyStorage) *MoneyService {
	return &MoneyService{
		moneyStorage: moneyStorage,
	}
}

func (ms *MoneyService) Withdraw(ctx context.Context, userId uuid.UUID, withdrawal *withdrawal.Withdrawal) error {
	if !orderhelper.ValidateOrderId(withdrawal.OrderId) {
		return exceptions.NewOrderBadIdFormatError()
	}
	if withdrawal.Sum <= 0 {
		return exceptions.NewBalanceBadAmountFormatError()
	}

	userBalance, err := ms.moneyStorage.GetBalance(ctx, userId)
	if err != nil {
		return err
	}
	if withdrawal.Sum > userBalance.Current {
		return exceptions.NewBalanceNotEnoughBalanceError()
	}

	existingWithdrawal, err := ms.moneyStorage.GetWithdrawal(ctx, withdrawal.Id)
	if existingWithdrawal != nil {
		return nil
	}
	if err != nil && !errors.Is(err, exceptions.NewWithdrawalNotFoundError()) {
		return err
	}

	tx, err := ms.moneyStorage.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = ms.moneyStorage.InsertWithdrawl(ctx, userId, withdrawal, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = ms.moneyStorage.ReduceBalance(ctx, userId, withdrawal.Sum, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

func (ms *MoneyService) AddAccrual(ctx context.Context, userId uuid.UUID, amount float64, trx *transaction.Trx) error {
	if amount <= 0 {
		return exceptions.NewBalanceBadAmountFormatError()
	}
	return ms.moneyStorage.AddBalance(ctx, userId, amount, trx)
}

func (ms *MoneyService) GetBalance(ctx context.Context, userId uuid.UUID) (*balance.Balance, error) {
	return ms.moneyStorage.GetBalance(ctx, userId)
}

func (ms *MoneyService) GetWithdrawals(ctx context.Context, userId uuid.UUID) ([]withdrawal.Withdrawal, error) {
	return ms.moneyStorage.GetWithdrawals(ctx, userId)
}
