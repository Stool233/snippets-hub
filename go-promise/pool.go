package promise

import (
	conc "github.com/sourcegraph/conc/pool"
)

var (
	defaultPool = newDefaultPool()
)

type Pool interface {
	Go(f func())
}

type wrapFunc func(f func())

func (wf wrapFunc) Go(f func()) {
	wf(f)
}

func newDefaultPool() Pool {
	return wrapFunc(func(f func()) {
		go f()
	})
}

func FromConcPool(p *conc.Pool) Pool {
	return wrapFunc(p.Go)
}
