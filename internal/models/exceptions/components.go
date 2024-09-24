package exceptions

import "errors"

var (
	ErrGracefullyShutDown = errors.New("gracefully shutdown")
)
