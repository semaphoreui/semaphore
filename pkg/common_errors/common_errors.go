package common_errors

import (
	"errors"
	"path"
	"runtime"
)

type UserVisibleError struct {
	Err error
}

func (e *UserVisibleError) Error() string { return e.Err.Error() }
func (e *UserVisibleError) Unwrap() error { return e.Err }

func NewUserError(err error) error {
	return &UserVisibleError{Err: err}
}

func NewUserErrorS(err string) error {
	return &UserVisibleError{Err: errors.New(err)}
}

var ErrInvalidSubscription = errors.New("has no active subscription")

func GetErrorContext() string {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	return path.Base(file) + ":" + path.Base(fn.Name()) + ":" + string(rune(line))
}
