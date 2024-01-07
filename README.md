# go-async

![test](https://github.com/ghosind/go-async/workflows/test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghosind/go-async)](https://goreportcard.com/report/github.com/ghosind/go-async)
[![codecov](https://codecov.io/gh/ghosind/go-async/branch/main/graph/badge.svg)](https://codecov.io/gh/ghosind/go-async)
![Version Badge](https://img.shields.io/github/v/release/ghosind/go-async)
![License Badge](https://img.shields.io/github/license/ghosind/go-async)
[![Go Reference](https://pkg.go.dev/badge/github.com/ghosind/go-async.svg)](https://pkg.go.dev/github.com/ghosind/go-async)

A tool collection that provided asynchronous workflow control utilities, inspired by `Promise` in the Javascript.

## Installation

```sh
go get -u github.com/ghosind/go-async
```

## Getting Started

We provided the `All` function to execute all the functions asynchronously. It'll return `-1` and a nil error if all functions are completed and no error return or panic. If some function returns an error or panic, it'll immediately return the index of the function and the error and send the cancel signal to all other functions.

```go
index, err := async.All(func (ctx context.Context) error) {
  return nil
}, func (ctx context.Context) error) {
  return nil
}/*, ...*/)
// index: -1
// err: <nil>

index, err := async.All(func (ctx context.Context) error) {
  return nil
}, func (ctx context.Context) error) {
  return errors.New("some error")
}/*, ...*/)
// index: 1
// err: Some error
```

If you do not want to terminate the execution of other functions if some function returns an error or panic, you can try the `AllCompleted` function. The `AllCompleted` function will return until all functions are finished or panic. It'll return a list of the function return values (error), and a boolean value to indicate any functions return error or panic.

```go
errors, ok := async.All(func (ctx context.Context) error) {
  return nil
}, func (ctx context.Context) error) {
  return errors.New("some error")
}/*, ...*/)
// errors: [<nil>, some error]
// ok: false
```

We also provided the `Race` function, it will return when a function returns or panics, and does not terminate other functions.

```go
index, err := async.Race(func (ctx context.Context) error {
  request.Get("https://example.com")
  return nil
}, func (ctx context.Context) error {
  time.Sleep(time.Second)
  return nil
})
// index = 0 if the request is finished within one second.
// index = 1 if the request is finished after one second.
```

To run all functions asynchronously but with the specified concurrency limitation, you can use the `Parallel` function. The `Parallel` function accepts a number that the concurrency limitation and the list of functions to run. The number of the concurrency limitation must be greater than or equal to 0, and it has the same effect as the `All` function if the number is 0.

```go
// Run 2 functions asynchronously at the time.
async.Parallel(2, func (ctx context.Context) error {
  // Do something
  return nil
}, func (ctx context.Context) error {
  // Do something
  return nil
}, func (ctx context.Context) error {
  // Do something
  return nil
}/* , ... */)
```

The `Parallel` will also be terminated if any function panics or returns an error. If you do not want to terminate the execution of other functions, you can try to use `ParallelCompleted`. The `ParallelCompleted` function will run all functions until all functions are finished. It will return the errors list and a boolean value to indicate whether any function errored.
