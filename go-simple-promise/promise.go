package gopromise

import (
	"fmt"
	"sync"
)

// 定义Promise的三种状态
const (
	Pending = iota
	Resolved
	Rejected
)

// Promise 结构体
type Promise struct {
	state    int
	result   interface{}
	mutex    sync.Mutex
	err      error
	doneChan chan struct{}
}

// NewPromise 创建一个新的Promise，接收一个执行函数
//
//	它提供了resolve、reject两个函数，给调用方去设置状态
func NewPromise(executor func(resolve func(interface{}), reject func(error))) *Promise {
	p := &Promise{
		state:    Pending,
		doneChan: make(chan struct{}),
	}

	go func() {

		// Recover from panic within the executor function
		defer func() {
			if r := recover(); r != nil {
				p.mutex.Lock()
				if p.state == Pending {
					p.state = Rejected
					p.err = fmt.Errorf("panic: %v", r)
					close(p.doneChan)
				}
				p.mutex.Unlock()
			}
		}()

		executor(
			func(result interface{}) { // resolve函数
				p.mutex.Lock()
				defer p.mutex.Unlock()
				if p.state != Pending {
					return
				}
				p.state = Resolved
				p.result = result
				close(p.doneChan)
			},
			func(err error) { // reject函数
				p.mutex.Lock()
				defer p.mutex.Unlock()
				if p.state != Pending {
					return
				}
				p.state = Rejected
				p.err = err
				close(p.doneChan)
			},
		)
	}()

	return p
}

// Then 注册resolved状态的处理函数
func (p *Promise) then(onResolved func(interface{}) interface{}, onRejected func(error)) *Promise {
	return NewPromise(func(resolve func(interface{}), reject func(error)) {
		<-p.doneChan
		p.mutex.Lock()
		resolvedValue := p.result
		rejectedError := p.err
		promiseState := p.state
		p.mutex.Unlock()

		if promiseState == Resolved {
			if onResolved != nil {
				// Call onResolved without holding the lock to avoid deadlocks
				// and allow other goroutines to access the Promise.
				// 在不持有锁的情况下，调用onResolved以避免死锁并允许其他例程访问Promise。
				result := onResolved(resolvedValue)
				resolve(result)
			} else {
				resolve(resolvedValue)
			}
		} else if promiseState == Rejected {
			if onRejected != nil {
				// Similarly, call onRejected without holding the lock.
				// 类似地，在不持有锁的情况下调用onRejected
				onRejected(rejectedError)
				// If onRejected is provided, it's assumed to handle the error,
				// so the new promise should be resolved to maintain the chain.
				// 如果提供了onRejected，则假定它将处理错误，因此应该解析新的Promise以维护链。
				resolve(nil)
			} else {
				reject(rejectedError)
			}
		}
	})
}

func (p *Promise) Then(onResolved func(interface{}) interface{}) *Promise {
	return p.then(onResolved, nil)
}

// Catch 注册rejected状态的处理函数
func (p *Promise) Catch(onRejected func(error)) *Promise {
	return p.then(nil, onRejected)
}
