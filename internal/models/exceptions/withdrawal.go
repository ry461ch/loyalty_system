package exceptions

import "fmt"

type WithdrawalStorageError struct {
	message string
}

func (wse WithdrawalStorageError) Error() string {
	return wse.message
}

func NewWithdrawalStorageError() error {
	return WithdrawalStorageError{
		message: "Withdrawal storage error occured",
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
		err:     NewWithdrawalStorageError(),
		message: "withdrawal not found",
	}
}
