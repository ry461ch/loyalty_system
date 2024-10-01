package exceptions

import "errors"

var (
	ErrBalanceBadAmountFormat = errors.New("bad balance amount format")
	ErrNotEnoughBalance       = errors.New("not enough money on the balance")
)
