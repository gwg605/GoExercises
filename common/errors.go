package common

import "errors"

var (
	ErrorFatal            = errors.New("fatal error")
	ErrorTimeout          = errors.New("timeout")
	ErrorNotFound         = errors.New("not found")
	ErrorInvalidState     = errors.New("invalid state")
	ErrorBadRequest       = errors.New("bad request")
	ErrorAlreadyExists    = errors.New("already exists")
	ErrorInvalidConfig    = errors.New("invalid config")
	ErrorNotImplemented   = errors.New("not implemented")
	ErrorOperationPending = errors.New("operation pending")
	ErrorNotApproved      = errors.New("not approved")
	ErrorTooMany          = errors.New("too many")
)
