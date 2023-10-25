package async

import (
	"errors"
	"fmt"
)

// executionContainer is a recoverable function container that provides recoverability from panic
// errors.
func executionContainer(fn func() error) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case error:
				err = x
			case string:
				err = errors.New(x)
			default:
				err = fmt.Errorf("%v", x)
			}
		}
	}()

	err = fn()

	return
}
