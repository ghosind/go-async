# go-async

![test](https://github.com/ghosind/go-async/workflows/test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghosind/go-async)](https://goreportcard.com/report/github.com/ghosind/go-async)
[![codecov](https://codecov.io/gh/ghosind/go-async/branch/main/graph/badge.svg)](https://codecov.io/gh/ghosind/go-async)
![Version Badge](https://img.shields.io/github/v/release/ghosind/go-async)
![License Badge](https://img.shields.io/github/license/ghosind/go-async)
[![Go Reference](https://pkg.go.dev/badge/github.com/ghosind/go-async.svg)](https://pkg.go.dev/github.com/ghosind/go-async)

简体中文 | [English](./README.md)

Golang异步工具集，启发自[JavaScript的`Promise`对象](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise)以及[Node.js async包](https://caolan.github.io/async/v3/).

## 安装与要求

执行下面的命令以安装本工具集，由于依赖的关系，对Golang的版本需要为1.18及以后版本。

```sh
go get -u github.com/ghosind/go-async
```

## 入门

### 允许的函数

本工具集中大部分的工具方法接受一个任意参数及返回值的函数并执行，但为了最好的效果，建议将第一个参数设置为`context.Context`以接收信号，并将最后一个返回值的类型设置为`error`用于告知工具方法是否发生了错误。

### 同时运行所有函数直到结束

`All`方法可以用于异步执行所有的函数，它将所有函数的返回值包装为一个2维的any类型切片并返回，且返回一个错误类型告知调用方法执行时是否有函数发生了错误。

在`All`方法执行时，一旦有任意一个函数发生错误，它将立即结束并返回该错误，并且通过传递的上下文发生取消信号。

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

若在执行过程中，即使有某个函数发生错误也不希望结束整体的运行，可以使用`AllCompleted`方法。`AllCompleted`方法会等待所有函数都执行完成或发生错误后，才会结束并返回所有函数的执行结果。

```go
out, err := async.All(func (ctx context.Context) (int, error) {
  return 0, nil
}, func (ctx context.Context) (string, error) {
  return "", errors.New("some error")
}/*, ...*/)
// out: [][]any{{0, nil}, {"", some error}}}
// err: function 1 error: some error
```

### 获取第一个结束的函数结果

在执行一系列函数时，若希望得到第一个执行完成的函数结果，可以使用`Race`方法。`Race`方法将异步执行所有函数，并在第一个执行结束的函数完成时结束该方法的执行并返回结果至调用方法。在任意一个函数执行结束后，`Race`方法将通过上下文发送一个取消信号至其它函数。

`Race`方法返回以下三个返回值：

- 2维执行结果列表；
- 第一个执行完毕的函数在参数列表中的位置；
- 第一个执行完毕的函数的错误。

```go
out, index, err := async.Race(func (ctx context.Context) (int, error) {
  request.Get("https://example.com")
  return 0, nil
}, func (ctx context.Context) (string, error) {
  time.Sleep(time.Second)
  return "test", nil
})
// 第一个函数比第二个函数快的情况下：
// out: []any{0, nil}, index: 0, err: nil
//
// 第二个函数比第二个函数快的情况下：
// out: []any{"test", nil}, index: 1, err: nil
```

### 在并发限制下运行

为了在并发限制条件下异步运行所有函数，可以使用`Parallel`方法。`Parallel`方法接受一个非负数的并发值用于限制运行时可以达到的最大并发数量，当该值为0时表示对并发数量没有限制，运行效果等同于使用`All`函数。

```go
// 在并发限制数为2的条件下运行
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

在某一个函数发生错误时，`Parallel`也会结束并发送取消信号至其它函数。若不希望结束其它函数的执行，可以使用`ParallelCompleted`代替，它将执行全部函数直到所有函数都完成或发生错误。

### 持续运行直到发生错误

通过`Forever`方法，可以持续运行某一函数直到该函数发生错误。`Forever`方法需要接受一个`ForeverFn`类型的方法用于执行，对于该类型的具体信息可以参考示例下方的描述。

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

`Forever`类型函数接受两个参数，第一个参数为从调用方法或上次调用传递来的上下文，第二个参数为为下次调用设置上下文的`next`函数。对于`ForeverFn`执行时，调用`next`函数设置上下文是一个可选的行为，且它只在第一次调用时生效。

### 自定义上下文

对于所有的工具方法，都有`XXXWithContext`版本（例如`AllWithContext`、`RaceWithContext`等），可以使用该版本传递自定义上下文。
