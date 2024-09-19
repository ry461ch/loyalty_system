package exceptions

import "fmt"

type BalanceStorageError struct {
	message string
}

func (bse BalanceStorageError) Error() string {
	return bse.message
}

func NewBalanceStorageError() error {
	return BalanceStorageError{
		message: "Balance storage error occured",
	}
}

type BalanceBadAmountFormatError struct {
	err     error
	message string
}

func (bbafe BalanceBadAmountFormatError) Error() string {
	return fmt.Sprintf("%v: %s", bbafe.err, bbafe.message)
}

func NewBalanceBadAmountFormatError() error {
	return BalanceBadAmountFormatError{
		err:     NewBalanceStorageError(),
		message: "bad balance amount format",
	}
}

type BalanceNotEnoughBalanceError struct {
	err     error
	message string
}

func (bnebe BalanceNotEnoughBalanceError) Error() string {
	return fmt.Sprintf("%v: %s", bnebe.err, bnebe.message)
}

func NewBalanceNotEnoughBalanceError() error {
	return BalanceNotEnoughBalanceError{
		err:     NewBalanceStorageError(),
		message: "not enough money on the balance",
	}
}
