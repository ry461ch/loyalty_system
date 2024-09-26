package moneyservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ry461ch/loyalty_system/internal/helpers/order"
	"github.com/ry461ch/loyalty_system/internal/helpers/transaction"
	"github.com/ry461ch/loyalty_system/internal/models/balance"
	"github.com/ry461ch/loyalty_system/internal/models/exceptions"
	"github.com/ry461ch/loyalty_system/internal/models/withdrawal"
)

type MoneyService struct {
	balanceStorage    BalanceStorage
	withdrawalStorage WithdrawalStorage
}

func NewMoneyService(balanceStorage BalanceStorage, withdrawalStorage WithdrawalStorage) *MoneyService {
	return &MoneyService{
		balanceStorage:    balanceStorage,
		withdrawalStorage: withdrawalStorage,
	}
}

func (ms *MoneyService) Withdraw(ctx context.Context, inputWithdrawal *withdrawal.Withdrawal) error {
	if !orderhelpers.ValidateOrderID(inputWithdrawal.OrderID) {
		return exceptions.ErrOrderBadIDFormat
	}
	if inputWithdrawal.Sum <= 0 {
		return exceptions.ErrBalanceBadAmountFormat
	}

	if inputWithdrawal.UserID == nil {
		return exceptions.ErrUserAuthentication
	}

	userBalance, err := ms.balanceStorage.GetBalance(ctx, *inputWithdrawal.UserID)
	if err != nil {
		return err
	}
	if inputWithdrawal.Sum > userBalance.Current {
		return exceptions.ErrNotEnoughBalance
	}

	if inputWithdrawal.ID == nil {
		inputWithdrawalID := uuid.New()
		inputWithdrawal.ID = &inputWithdrawalID
	}
	existingWithdrawal, err := ms.withdrawalStorage.GetWithdrawal(ctx, *inputWithdrawal.ID)
	if existingWithdrawal != nil {
		return nil
	}
	if err != nil && !errors.Is(err, exceptions.ErrWithdrawalNotFound) {
		return err
	}

	tx, err := ms.withdrawalStorage.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = ms.withdrawalStorage.InsertWithdrawal(ctx, inputWithdrawal, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = ms.balanceStorage.ReduceBalance(ctx, *inputWithdrawal.UserID, inputWithdrawal.Sum, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

func (ms *MoneyService) GetBalance(ctx context.Context, userID uuid.UUID) (*balance.Balance, error) {
	return ms.balanceStorage.GetBalance(ctx, userID)
}

func (ms *MoneyService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]withdrawal.Withdrawal, error) {
	return ms.withdrawalStorage.GetWithdrawals(ctx, userID)
}

func (ms *MoneyService) AddAccrual(ctx context.Context, userID uuid.UUID, amount float64, trx *transaction.Trx) error {
	if amount <= 0 {
		return exceptions.ErrBalanceBadAmountFormat
	}
	return ms.balanceStorage.AddBalance(ctx, userID, amount, trx)
}
