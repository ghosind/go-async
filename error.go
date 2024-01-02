package async

import "errors"

var (
	// ErrContextCanceled to indicate the context was canceled or timed out.
	ErrContextCanceled error = errors.New("context canceled")
)
