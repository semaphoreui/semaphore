package common_errors

import (
	"errors"
	"path"
	"runtime"
)

var ErrInvalidSubscription = errors.New("has no active subscription")

func GetErrorContext() string {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	return path.Base(file) + ":" + path.Base(fn.Name()) + ":" + string(rune(line))
}
