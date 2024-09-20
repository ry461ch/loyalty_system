package exceptions

import "fmt"

type UserError struct {
	message string
}

func (use UserError) Error() string {
	return use.message
}

func NewUserError() error {
	return UserError{
		message: "user error occured",
	}
}

type UserNotFoundError struct {
	err     error
	message string
}

func (unfe UserNotFoundError) Error() string {
	return fmt.Sprintf("%v: %s", unfe.err, unfe.message)
}

func NewUserNotFoundError() error {
	return UserNotFoundError{
		err:     NewUserError(),
		message: "user not found",
	}
}

type UserAuthenticationError struct {
	err     error
	message string
}

func (unfe UserAuthenticationError) Error() string {
	return fmt.Sprintf("%v: %s", unfe.err, unfe.message)
}

func NewUserAuthenticationError() error {
	return UserAuthenticationError{
		err:     NewUserError(),
		message: "user unauthorized",
	}
}

type UserConflictError struct {
	err     error
	message string
}

func (unfe UserConflictError) Error() string {
	return fmt.Sprintf("%v: %s", unfe.err, unfe.message)
}

func NewUserConflictError() error {
	return UserConflictError{
		err:     NewUserError(),
		message: "user already exists",
	}
}

type UserBadFormatError struct {
	err     error
	message string
}

func (ubfe UserBadFormatError) Error() string {
	return fmt.Sprintf("%v: %s", ubfe.err, ubfe.message)
}

func NewUserBadFormatError() error {
	return UserBadFormatError{
		err:     NewUserError(),
		message: "user bad data format",
	}
}
