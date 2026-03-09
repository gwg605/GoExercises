package base

import "errors"

var (
	ErrorFatal            = errors.New("fatal error")
	ErrorBadRequest       = errors.New("bad request")
	ErrorTimeout          = errors.New("timeout")
	ErrorNotFound         = errors.New("not found")
	ErrorInvalidState     = errors.New("invalid state")
	ErrorAlreadyExists    = errors.New("already exists")
	ErrorInvalidConfig    = errors.New("invalid config")
	ErrorNotImplemented   = errors.New("not implemented")
	ErrorOperationPending = errors.New("operation pending")
	ErrorTooMany          = errors.New("too many")
	ErrorAborted          = errors.New("aborted")
	ErrorOutOfRange       = errors.New("out of range")
)
