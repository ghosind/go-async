package async

import (
	"context"
	"testing"

	"github.com/ghosind/go-assert"
)

func TestGetContext(t *testing.T) {
	a := assert.New(t)

	todoCtx := context.TODO()
	ctx := getContext(todoCtx)
	a.EqualNow(ctx, todoCtx)

	//lint:ignore SA1012 for test case only
	ctx = getContext(nil)
	a.NotNilNow(ctx)
	a.NotEqualNow(ctx, todoCtx)
}
