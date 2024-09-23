package exceptions

import "errors"

var (
	ErrWithdrawalNotFound  = errors.New("withdrawal not found")
	ErrWithdrawalBadFormat = errors.New("withdrawal bad format")
)
