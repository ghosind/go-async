package async

import (
	"context"
	"sync"
	"sync/atomic"
)

// builtinPool is the Parallelers pool for built-in functions.
var builtinPool sync.Pool = sync.Pool{
	New: func() any {
		return new(Paralleler)
	},
}

// Paralleler is a tool to run the tasks with the specific concurrency, default no concurrency
// limitation.
type Paralleler struct {
	concurrency int
	ctx         context.Context
	locker      sync.Mutex
	tasks       []AsyncFn
}

// WithConcurrency sets the number of concurrency limitation.
func (p *Paralleler) WithConcurrency(concurrency int) *Paralleler {
	if concurrency < 0 {
		panic(ErrInvalidConcurrency)
	}

	p.concurrency = concurrency

	return p
}

// WithContext sets the context that passes to the tasks.
func (p *Paralleler) WithContext(ctx context.Context) *Paralleler {
	p.ctx = ctx

	return p
}

// Add adds the functions into the pending tasks list.
func (p *Paralleler) Add(funcs ...AsyncFn) *Paralleler {
	validateAsyncFuncs(funcs...)

	p.locker.Lock()
	defer p.locker.Unlock()

	p.tasks = append(p.tasks, funcs...)

	return p
}

// Clear clears the paralleler's pending tasks list.
func (p *Paralleler) Clear() *Paralleler {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.tasks = nil

	return p
}

// Run runs the tasks in the paralleler's pending list, it'll clear the pending list and return
// the results of the tasks.
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

	go p.runTasks(ctx, ch, tasks, true)

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

// RunCompleted runs the tasks in the paralleler's pending list until all functions are finished,
// it'll clear the pending list and return the results of the tasks.
func (p *Paralleler) RunCompleted() ([][]any, error) {
	tasks := p.getTasks()
	out := make([][]any, len(tasks))
	if len(tasks) == 0 {
		return out, nil
	}

	errs := make([]error, len(tasks))
	errNum := atomic.Int32{}
	parent := getContext(p.ctx)
	ctx, canFunc := context.WithCancel(parent)
	defer canFunc()

	ch := make(chan executeResult, len(tasks))

	go p.runTasks(ctx, ch, tasks, false)

	for finished := 0; finished < len(tasks); finished++ {
		ret := <-ch
		out[ret.Index] = ret.Out
		if ret.Error != nil {
			errs[ret.Index] = ret.Error
			errNum.Add(1)
		}
	}

	if errNum.Load() == 0 {
		return out, nil
	}

	return out, convertErrorListToExecutionErrors(errs, int(errNum.Load()))
}

// getConcurrencyChan creates and returns a concurrency controlling channel by the specific number
// of the concurrency limitation.
func (p *Paralleler) getConcurrencyChan() chan empty {
	var conch chan empty

	if p.concurrency > 0 {
		conch = make(chan empty, p.concurrency)
	}

	return conch
}

// getTasks returns the tasks from the pending list, and clear the pending list to receiving new
// tasks.
func (p *Paralleler) getTasks() []AsyncFn {
	p.locker.Lock()

	tasks := p.tasks
	p.tasks = nil

	p.locker.Unlock()

	return tasks
}

// runTasks runs the tasks with the concurrency limitation.
func (p *Paralleler) runTasks(
	ctx context.Context,
	resCh chan executeResult,
	tasks []AsyncFn,
	exitWhenDone bool,
) {
	conch := p.getConcurrencyChan()

	for i := 0; i < len(tasks); i++ {
		if conch != nil {
			conch <- empty{}
		}

		go p.runTask(ctx, i, tasks[i], conch, resCh, exitWhenDone)
	}
}

// runTask runs the task function, and
func (p *Paralleler) runTask(
	ctx context.Context,
	n int,
	fn AsyncFn,
	conch chan empty,
	ch chan executeResult,
	exitWhenDone bool,
) {
	childCtx, childCanFunc := context.WithCancel(ctx)
	defer childCanFunc()

	ret, err := invokeAsyncFn(fn, childCtx, nil)

	if conch != nil {
		<-conch
	}

	if !exitWhenDone {
		ch <- executeResult{
			Index: n,
			Error: err,
			Out:   ret,
		}
	} else {
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
}
