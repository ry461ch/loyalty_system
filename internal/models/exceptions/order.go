package exceptions

import "errors"

var (
	ErrOrderBadIDFormat         = errors.New("bad order id format")
	ErrOrderBadStatusFormat     = errors.New("bad order status format")
	ErrOrderConflictSameUser    = errors.New("order already exists with same user")
	ErrOrderConflictAnotherUser = errors.New("order already exists with another user")
	ErrOrderNotFound            = errors.New("order not found")
)
