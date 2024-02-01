# go-async

![test](https://github.com/ghosind/go-async/workflows/test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghosind/go-async)](https://goreportcard.com/report/github.com/ghosind/go-async)
[![codecov](https://codecov.io/gh/ghosind/go-async/branch/main/graph/badge.svg)](https://codecov.io/gh/ghosind/go-async)
![Version Badge](https://img.shields.io/github/v/release/ghosind/go-async)
![License Badge](https://img.shields.io/github/license/ghosind/go-async)
[![Go Reference](https://pkg.go.dev/badge/github.com/ghosind/go-async.svg)](https://pkg.go.dev/github.com/ghosind/go-async)

A tool collection that provided asynchronous workflow control utilities, inspired by [JavaScript `Promise` Object](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise) and [Node.js async package](https://caolan.github.io/async/v3/).

## Installation and Requirement

Run the following command to install the library, and this library requires Go 1.18 and later versions.

```sh
go get -u github.com/ghosind/go-async
```

## Getting Started

### The function to run

The most of utility functions of this library accept any type of function to run, you can set the parameters and the return values as any type and any number of return values that you want. However, for best practice, we recommend you to set the first parameter as `context.Context` to receive the signals and make the type of the last return value as an error to let the utilities know whether an error happened or not.

### Run all functions until they are finished

`All` function can help you to execute all the functions asynchronously. It'll wrap all return values to a two-dimensional `any` type slice and return it if all functions are completed and no error returns or panic.

If any function returns an error or panics, the `All` function will terminate immediately and return the error. It'll also send a cancel signal to other uncompleted functions by context if they accept a context by their first parameter.

```go
out, err := async.All(func (ctx context.Context) (int, error)) {
  return 0, nil
}, func (ctx context.Context) (string, error)) {
  return "hello", nil
}/*, ...*/)
// out: [][]any{{0, nil}, {"hello", nil}}
// err: <nil>

out, err := async.All(func (ctx context.Context) (int, error)) {
  return 0, nil
}, func (ctx context.Context) (string, error)) {
  return "", errors.New("some error")
}/*, ...*/)
// out: nil
// err: some error
```

If you do not want to terminate the execution when some function returns an error or panic, you can try the `AllCompleted` function. The `AllCompleted` function executes until all functions are finished or panic. It'll return a list of the function return values, and an error to indicate whether any functions return error or panic.

```go
out, err := async.All(func (ctx context.Context) (int, error) {
  return 0, nil
}, func (ctx context.Context) (string, error) {
  return "", errors.New("some error")
}/*, ...*/)
// out: [][]any{{0, nil}, {"", some error}}}
// err: function 1 error: some error
```

### Get first output

If you want to run a list of functions and get the output of the first finish function, you can try the `Race` function. The `Race` function will run all functions asynchronously, and return when a function is finished or panicked.

The `Race` function returns three values:

- 1st value: an output list of the first finish function.
- 2nd value: the index of the first finish function.
- 3rd value: the execution error that from the first finish or panic function.

```go
out, index, err := async.Race(func (ctx context.Context) (int, error) {
  request.Get("https://example.com")
  return 0, nil
}, func (ctx context.Context) (string, error) {
  time.Sleep(time.Second)
  return "test", nil
})
// If the first function faster than the second one:
// out: []any{0, nil}, index: 0, err: nil
//
// Otherwise:
// out: []any{"test", nil}, index: 1, err: nil
```

### Run all functions with concurrency limit

To run all functions asynchronously but with the specified concurrency limitation, you can use the `Parallel` function. The `Parallel` function accepts a number that the concurrency limitation and the list of functions to run. The number of the concurrency limitation must be greater than or equal to 0, and it has the same effect as the `All` function if the number is 0.

```go
// Run 2 functions asynchronously at the time.
out, err := async.Parallel(2, func (ctx context.Context) (int, error) {
  // Do something
  return 1, nil
}, func (ctx context.Context) (string, error) {
  // Do something
  return "hello", nil
}, func (ctx context.Context) error {
  // Do something
  return nil
}/* , ... */)
// out: [][]any{{1, nil}, {"hello", nil}, {nil}}
// err: nil
```

The `Parallel` will also be terminated if any function panics or returns an error. If you do not want to terminate the execution of other functions, you can try to use `ParallelCompleted`. The `ParallelCompleted` function will run all functions until all functions are finished. It will return the errors list and a boolean value to indicate whether any function errored.

### Run a function forever until it returns an error or panic

For `Forever` function, you can use it to run a function forever until it returns an error or panics. You need to run the `Forever` function with a `ForeverFn` type function, and you can see more information about `ForeverFn` after the following example.

```go
err := async.Forever(func(ctx context.Context, next func(context.Context)) error {
  v, ok := ctx.Value("key")
  if ok {
    vi := v.(int)
    if vi == 5 {
      return errors.New("finish")
    }

    fmt.Printf("value: %d\n", vi)

    next(context.WithValue(ctx, "key", vi + 1))
  } else {
    next(context.WithValue(ctx, "key", 1))
  }
})
fmt.Printf("err: %v\n", err)
// value: 1
// value: 2
// value: 3
// value: 4
// err: finish
```

The `ForeverFn` accepts two parameters, the first one is a context from the caller or the last invocation. The second parameter of `ForeverFn` is a function to set the context that passes to the next invocation. For `ForeverFn`, it is an optional behavior to call the `next` function, and only the first invoke will work.

### Customize context

For all utility functions, you can use the `XXXWithContext` function (like `AllWithContext`, `RaceWithContext`, ...) to set the context by yourself.
