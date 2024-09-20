package exceptions

import "fmt"

type WithdrawalError struct {
	message string
}

func (wse WithdrawalError) Error() string {
	return wse.message
}

func NewWithdrawalError() error {
	return WithdrawalError{
		message: "Withdrawal error occured",
	}
}

type WithdrawalNotFoundError struct {
	err     error
	message string
}

func (wnfe WithdrawalNotFoundError) Error() string {
	return fmt.Sprintf("%v: %s", wnfe.err, wnfe.message)
}

func NewWithdrawalNotFoundError() error {
	return WithdrawalNotFoundError{
		err:     NewWithdrawalError(),
		message: "withdrawal not found",
	}
}

type WithdrawalBadFormatError struct {
	err     error
	message string
}

func (wbfe WithdrawalBadFormatError) Error() string {
	return fmt.Sprintf("%v: %s", wbfe.err, wbfe.message)
}

func NewWithdrawalBadFormatError() error {
	return WithdrawalBadFormatError{
		err:     NewWithdrawalError(),
		message: "withdrawal bad format",
	}
}
