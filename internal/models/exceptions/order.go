package exceptions

import "fmt"

type OrderStorageError struct {
	message string
}

func (ose OrderStorageError) Error() string {
	return ose.message
}

func NewOrderStorageError() error {
	return OrderStorageError{
		message: "Order storage error occured",
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
		err:     NewOrderStorageError(),
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
		err:     NewOrderStorageError(),
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
		err:     NewOrderStorageError(),
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
		err:     NewOrderStorageError(),
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
		err:     NewOrderStorageError(),
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
		err:     NewOrderStorageError(),
		message: "order not found",
	}
}
