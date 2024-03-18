package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

//go:noinline
func Assertf(condition bool, format string, args ...any) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Errorf(fmt.Sprintf("%v(%v): ", filepath.Base(file), line)+format, args...))
	}
}
