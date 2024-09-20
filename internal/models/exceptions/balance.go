package exceptions

import "fmt"

type BalanceError struct {
	message string
}

func (bse BalanceError) Error() string {
	return bse.message
}

func NewBalanceError() error {
	return BalanceError{
		message: "Balance error occured",
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
		err:     NewBalanceError(),
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
		err:     NewBalanceError(),
		message: "not enough money on the balance",
	}
}
