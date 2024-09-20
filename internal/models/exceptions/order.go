package exceptions

import "fmt"

type OrderError struct {
	message string
}

func (ose OrderError) Error() string {
	return ose.message
}

func NewOrderError() error {
	return OrderError{
		message: "Order error occured",
	}
}

type OrderBadIdFormatError struct {
	err     error
	message string
}

func (obife OrderBadIdFormatError) Error() string {
	return fmt.Sprintf("%v: %s", obife.err, obife.message)
}

func NewOrderBadIdFormatError() error {
	return OrderBadIdFormatError{
		err:     NewOrderError(),
		message: "bad order id format",
	}
}

type OrderBadStatusFormatError struct {
	err     error
	message string
}

func (obsfe OrderBadStatusFormatError) Error() string {
	return fmt.Sprintf("%v: %s", obsfe.err, obsfe.message)
}

func NewOrderBadStatusFormatError() error {
	return OrderBadStatusFormatError{
		err:     NewOrderError(),
		message: "bad order status format",
	}
}

type OrderConflictError struct {
	err     error
	message string
}

func (oce OrderConflictError) Error() string {
	return fmt.Sprintf("%v: %s", oce.err, oce.message)
}

func NewOrderConflictError() error {
	return OrderConflictError{
		err:     NewOrderError(),
		message: "order already exists",
	}
}

type OrderConflictSameUserError struct {
	err     error
	message string
}

func (ocsue OrderConflictSameUserError) Error() string {
	return fmt.Sprintf("%v: %s", ocsue.err, ocsue.message)
}

func NewOrderConflictSameUserError() error {
	return OrderConflictSameUserError{
		err:     NewOrderError(),
		message: "order already exists with same user",
	}
}

type OrderConflictAnotherUserError struct {
	err     error
	message string
}

func (ocaue OrderConflictAnotherUserError) Error() string {
	return fmt.Sprintf("%v: %s", ocaue.err, ocaue.message)
}

func NewOrderConflictAnotherUserError() error {
	return OrderConflictAnotherUserError{
		err:     NewOrderError(),
		message: "order already exists with another user",
	}
}

type OrderNotFoundError struct {
	err     error
	message string
}

func (onfe OrderNotFoundError) Error() string {
	return fmt.Sprintf("%v: %s", onfe.err, onfe.message)
}

func NewOrderNotFoundError() error {
	return OrderNotFoundError{
		err:     NewOrderError(),
		message: "order not found",
	}
}
