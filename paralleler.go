package async

import (
	"context"
	"sync"
)

type Paralleler struct {
	concurrency int
	ctx         context.Context
	locker      sync.Mutex
	tasks       []AsyncFn
}

func (p *Paralleler) WithConcurrency(concurrency int) *Paralleler {
	if concurrency < 0 {
		panic(ErrInvalidConcurrency)
	}

	p.concurrency = concurrency

	return p
}

func (p *Paralleler) WithContext(ctx context.Context) *Paralleler {
	p.ctx = ctx

	return p
}

func (p *Paralleler) Add(funcs ...AsyncFn) *Paralleler {
	validateAsyncFuncs(funcs...)

	p.locker.Lock()
	defer p.locker.Unlock()

	p.tasks = append(p.tasks, funcs...)

	return p
}

func (p *Paralleler) Run() ([][]any, error) {
	tasks := p.getTasks()
	out := make([][]any, len(tasks))
	if len(tasks) == 0 {
		return out, nil
	}

	parent := getContext(p.ctx)
	ctx, canFunc := context.WithCancel(parent)
	defer canFunc()

	ch := make(chan executeResult, len(tasks))

	go p.runTasks(ctx, ch, tasks)

	finished := 0
	for finished < len(tasks) {
		select {
		case <-parent.Done():
			return out, ErrContextCanceled
		case ret := <-ch:
			out[ret.Index] = ret.Out
			if ret.Error != nil {
				return out, &executionError{
					index: ret.Index,
					err:   ret.Error,
				}
			}
			finished++
		}
	}

	return out, nil
}

func (p *Paralleler) getConcurrencyChan() chan empty {
	var conch chan empty

	if p.concurrency > 0 {
		conch = make(chan empty, p.concurrency)
	}

	return conch
}

func (p *Paralleler) getTasks() []AsyncFn {
	p.locker.Lock()

	tasks := p.tasks
	p.tasks = nil

	p.locker.Unlock()

	return tasks
}

func (p *Paralleler) runTasks(ctx context.Context, resCh chan executeResult, tasks []AsyncFn) {
	conch := p.getConcurrencyChan()

	for i := 0; i < len(tasks); i++ {
		if conch != nil {
			conch <- empty{}
		}

		go p.runTask(ctx, i, tasks[i], conch, resCh)
	}
}

func (p *Paralleler) runTask(
	ctx context.Context,
	n int,
	fn AsyncFn,
	conch chan empty,
	ch chan executeResult,
) {
	childCtx, childCanFunc := context.WithCancel(ctx)
	defer childCanFunc()

	ret, err := invokeAsyncFn(fn, childCtx, nil)

	if conch != nil {
		<-conch
	}

	select {
	case <-ctx.Done():
		return
	default:
		ch <- executeResult{
			Index: n,
			Error: err,
			Out:   ret,
		}
	}
}
