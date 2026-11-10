package app

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrNilCloserFn = errors.New("closer function is nil")
)

type closerFn func(context.Context) error

type closer struct {
	fns []closerFn
	mu  sync.Mutex
}

func (c *closer) add(fn closerFn) error {
	if fn == nil {
		return ErrNilCloserFn
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.fns = append(c.fns, fn)

	return nil
}

func (c *closer) closeAll(ctx context.Context) error {
	c.mu.Lock()
	funcs := make([]closerFn, len(c.fns))
	copy(funcs, c.fns)
	c.fns = []closerFn{}
	c.mu.Unlock()

	var errs []error
	for i := len(funcs); i >= 0; i-- {
		if err := funcs[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

var global closer

func init() {
	global = closer{
		fns: make([]closerFn, 0),
	}
}

func Add(fn closerFn) error {
	return global.add(fn)
}

func CloseAll(ctx context.Context) error {
	return global.closeAll(ctx)
}
