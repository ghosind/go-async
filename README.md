# go-async

![test](https://github.com/ghosind/go-async/workflows/test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghosind/go-async)](https://goreportcard.com/report/github.com/ghosind/go-async)
[![codecov](https://codecov.io/gh/ghosind/go-async/branch/main/graph/badge.svg)](https://codecov.io/gh/ghosind/go-async)
![Version Badge](https://img.shields.io/github/v/release/ghosind/go-async)
![License Badge](https://img.shields.io/github/license/ghosind/go-async)
[![Go Reference](https://pkg.go.dev/badge/github.com/ghosind/go-async.svg)](https://pkg.go.dev/github.com/ghosind/go-async)

English | [简体中文](./README-CN.md)

It's a powerful asynchronous utility library inspired by [JavaScript `Promise` Object](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise) and [Node.js async package](https://caolan.github.io/async/v3/).

## Installation and Requirement

Run the following command to install this library, and Go 1.18 and later versions required.

```sh
go get -u github.com/ghosind/go-async
```

And then, import the library into your own code.

```go
import "github.com/ghosind/go-async"
```

> [!NOTE]
> This library is not stable yet, anything may change in the later versions.

## Getting Started

For the following example, it runs the functions concurrently and returns the return values until all functions have been completed.

```go
out, err := async.All(func (ctx context.Context) (int, error) {
  return 0, nil
}, func () (string, error)) {
  time.Sleep(100 * time.Millisecond)
  return "hello", nil
})
// out: [][]any{{0, <nil>}, {"hello", <nil>}}
// err: <nil>
```

There are over 10 asynchronous control flow functions available, please visit [Go Reference](https://pkg.go.dev/github.com/ghosind/go-async) to see the documentation and examples.

## The function to run

The most of utility functions of this library accept any type of function to run, you can set the parameters and the return values as any type and any number of return values that you want. However, for best practice, we recommend you to set the first parameter as `context.Context` to receive the signals and make the type of the last return value as an error to let the utilities know whether an error happened or not.

## Customize context

For all functions, you can use the `XXXWithContext` function (like `AllWithContext`, `RaceWithContext`, ...) to set the context by yourself.

## Available Functions

- [`All`](https://pkg.go.dev/github.com/ghosind/go-async#All)
- [`AllCompleted`](https://pkg.go.dev/github.com/ghosind/go-async#AllCompleted)
- [`Fallback`](https://pkg.go.dev/github.com/ghosind/go-async#Fallback)
- [`Forever`](https://pkg.go.dev/github.com/ghosind/go-async#Forever)
- [`Parallel`](https://pkg.go.dev/github.com/ghosind/go-async#Parallel)
- [`ParallelCompleted`](https://pkg.go.dev/github.com/ghosind/go-async#ParallelCompleted)
- [`Race`](https://pkg.go.dev/github.com/ghosind/go-async#Race)
- [`Retry`](https://pkg.go.dev/github.com/ghosind/go-async#Retry)
- [`Seq`](https://pkg.go.dev/github.com/ghosind/go-async#Seq)
- [`SeqGroups`](https://pkg.go.dev/github.com/ghosind/go-async#SeqGroups)
- [`Series`](https://pkg.go.dev/github.com/ghosind/go-async#Series)
- [`Times`](https://pkg.go.dev/github.com/ghosind/go-async#Times)
- [`TimesLimit`](https://pkg.go.dev/github.com/ghosind/go-async#TimesLimit)
- [`TimesSeries`](https://pkg.go.dev/github.com/ghosind/go-async#TimesSeries)
- [`Until`](https://pkg.go.dev/github.com/ghosind/go-async#Until)
- [`While`](https://pkg.go.dev/github.com/ghosind/go-async#While)

## License

The library published under MIT License, please see license file for more details.
