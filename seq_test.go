package async_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ghosind/go-assert"
	"github.com/ghosind/go-async"
)

func TestSeq(t *testing.T) {
	a := assert.New(t)

	out, err := async.Seq(func() int {
		return 1
	}, func(n int) int {
		return n + 1
	})
	a.NilNow(err)
	a.EqualNow(out, []any{2})
}

func TestSeqWithFailure(t *testing.T) {
	a := assert.New(t)
	expectedErr := errors.New("expected error")

	_, err := async.Seq(func() error {
		return expectedErr
	}, func(err error) {
		a.FailNow()
	})
	a.NotNilNow(err)
	a.ContainsStringNow(err.Error(), expectedErr.Error())
}

func TestSeqWithContext(t *testing.T) {
	a := assert.New(t)

	out, err := async.SeqWithContext(context.Background(), func() int {
		return 1
	}, func(n int) int {
		return n + 1
	})
	a.NilNow(err)
	a.EqualNow(out, []any{2})
}

func ExampleSeq() {
	out, err := async.Seq(func() int {
		return 1
	}, func(n int) int {
		return n + 1
	})
	fmt.Println(out)
	fmt.Println(err)
	// Output:
	// [2]
	// <nil>
}
