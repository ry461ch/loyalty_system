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
	balanceStorage    storage.BalanceStorage
	withdrawalStorage storage.WithdrawalStorage
}

func NewMoneyService(balanceStorage storage.BalanceStorage, withdrawalStorage storage.WithdrawalStorage) *MoneyService {
	return &MoneyService{
		balanceStorage:    balanceStorage,
		withdrawalStorage: withdrawalStorage,
	}
}

func (ms *MoneyService) Withdraw(ctx context.Context, inputWithdrawal *withdrawal.Withdrawal) error {
	if !orderhelper.ValidateOrderId(inputWithdrawal.OrderId) {
		return exceptions.NewOrderBadIdFormatError()
	}
	if inputWithdrawal.Sum <= 0 {
		return exceptions.NewBalanceBadAmountFormatError()
	}

	if inputWithdrawal.UserId == nil {
		return exceptions.NewUserAuthenticationError()
	}

	userBalance, err := ms.balanceStorage.GetBalance(ctx, *inputWithdrawal.UserId)
	if err != nil {
		return err
	}
	if inputWithdrawal.Sum > userBalance.Current {
		return exceptions.NewBalanceNotEnoughBalanceError()
	}

	if inputWithdrawal.Id == nil {
		inputWithdrawalId := uuid.New()
		inputWithdrawal.Id = &inputWithdrawalId
	}
	existingWithdrawal, err := ms.withdrawalStorage.GetWithdrawal(ctx, *inputWithdrawal.Id)
	if existingWithdrawal != nil {
		return nil
	}
	if err != nil && !errors.Is(err, exceptions.NewWithdrawalNotFoundError()) {
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

	err = ms.balanceStorage.ReduceBalance(ctx, *inputWithdrawal.UserId, inputWithdrawal.Sum, tx)
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
	return ms.balanceStorage.AddBalance(ctx, userId, amount, trx)
}

func (ms *MoneyService) GetBalance(ctx context.Context, userId uuid.UUID) (*balance.Balance, error) {
	return ms.balanceStorage.GetBalance(ctx, userId)
}

func (ms *MoneyService) GetWithdrawals(ctx context.Context, userId uuid.UUID) ([]withdrawal.Withdrawal, error) {
	return ms.withdrawalStorage.GetWithdrawals(ctx, userId)
}
