package async

import (
	"context"
	"time"
)

const (
	defaultRetryTimes    int = 5
	defaultRetryInterval int = 0
)

type RetryOptions struct {
	// Times is the number of attempts to make before giving up, the default is 5.
	Times int
	// Interval is the time to wait between retries in milliseconds, the default is 0.
	Interval int
	// IntervalFunc is the function to calculate the time to wait between retries in milliseconds, it
	// accepts an int value to indicate the retry count.
	IntervalFunc func(int) int
	// ErrorFilter is a function that is invoked on an error result. Retry will continue the retry
	// attempts if it returns true, and it will abort the workflow and return the current attempt's
	// result and error if it returns false.
	ErrorFilter func(error) bool
}

// Retry attempts to get a successful response from the function with no more than the specific
// retry times before returning an error. If the task is successful, it will return the result of
// the successful task. If all attempts fail, it will return the result and the error of the final
// attempt.
func Retry(fn AsyncFn, opts ...RetryOptions) ([]any, error) {
	return retry(context.Background(), fn, opts...)
}

// RetryWithContext runs the function with the specified context, and attempts to get a successful
// response from the function with no more than the specific retry times before returning an error.
// If the task is successful, it will return the result of the successful task. If all attempts
// fail, it will return the result and the error of the final attempt.
func RetryWithContext(ctx context.Context, fn AsyncFn, opts ...RetryOptions) ([]any, error) {
	return retry(ctx, fn, opts...)
}

// retry runs the function and attempts to get a successful response from the function with no more
// than the specific retry times before returning an error.
func retry(parent context.Context, fn AsyncFn, opts ...RetryOptions) (out []any, err error) {
	validateAsyncFuncs(fn)
	ctx := getContext(parent)
	opt := getRetryOption(opts...)

	for i := 1; i <= opt.Times; i++ {
		out, err = invokeAsyncFn(fn, ctx, nil)
		if err == nil {
			return
		} else if opt.ErrorFilter != nil && !opt.ErrorFilter(err) {
			return
		}

		if i != opt.Times {
			interval := opt.Interval
			if opt.IntervalFunc != nil {
				interval = opt.IntervalFunc(i)
			}

			if interval != 0 {
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}
	}

	return
}

// getRetryOption gets the retry option by the customize option or the default values.
func getRetryOption(opts ...RetryOptions) RetryOptions {
	opt := RetryOptions{}
	if len(opts) > 0 {
		opt.Interval = opts[0].Interval
		opt.Times = opts[0].Times
		opt.IntervalFunc = opts[0].IntervalFunc
		opt.ErrorFilter = opts[0].ErrorFilter
	}

	if opt.Interval <= 0 {
		opt.Interval = defaultRetryInterval
	}
	if opt.Times <= 0 {
		opt.Times = defaultRetryTimes
	}

	return opt
}
