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

安装后在Go项目源码中导入。

```go
import "github.com/ghosind/go-async"
```

## 入门

下面的代码中，通过`All`函数并发执行函数直到全部执行完成，并返回它们的返回结果。

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

本工具集中包含有超过十个异步控制方法，请前往[Go Reference](https://pkg.go.dev/github.com/ghosind/go-async)获取详细的文档及示例。

## 允许的函数

本工具集中大部分的工具方法接受一个任意参数及返回值的函数并执行，但为了最好的效果，建议将第一个参数设置为`context.Context`以接收信号，并将最后一个返回值的类型设置为`error`用于告知工具方法是否发生了错误。

## 自定义上下文

对于所有的工具方法，都有`XXXWithContext`版本（例如`AllWithContext`、`RaceWithContext`等），可以使用该版本传递自定义上下文。

## 可用的工具方法

- [`All`](https://pkg.go.dev/github.com/ghosind/go-async#All)
- [`AllCompleted`](https://pkg.go.dev/github.com/ghosind/go-async#AllCompleted)
- [`Forever`](https://pkg.go.dev/github.com/ghosind/go-async#Forever)
- [`Parallel`](https://pkg.go.dev/github.com/ghosind/go-async#Parallel)
- [`ParallelCompleted`](https://pkg.go.dev/github.com/ghosind/go-async#ParallelCompleted)
- [`Race`](https://pkg.go.dev/github.com/ghosind/go-async#Race)
- [`Retry`](https://pkg.go.dev/github.com/ghosind/go-async#Retry)
- [`Seq`](https://pkg.go.dev/github.com/ghosind/go-async#Seq)
- [`Series`](https://pkg.go.dev/github.com/ghosind/go-async#Series)
- [`Times`](https://pkg.go.dev/github.com/ghosind/go-async#Times)
- [`TimesLimit`](https://pkg.go.dev/github.com/ghosind/go-async#TimesLimit)
- [`TimesSeries`](https://pkg.go.dev/github.com/ghosind/go-async#TimesSeries)
- [`Until`](https://pkg.go.dev/github.com/ghosind/go-async#Until)
- [`While`](https://pkg.go.dev/github.com/ghosind/go-async#While)

## 许可

本项目通过MIT许可发布，请通过license文件获取该许可的详细信息。
