package gopromise

import (
	"errors"
	"testing"
	"time"
)

// 测试Promise成功解决
func TestPromiseResolved(t *testing.T) {
	resolvedValue := "success"
	p := NewPromise(func(resolve func(interface{}), reject func(error)) {
		resolve(resolvedValue)
	})

	<-p.doneChan // 等待Promise解决
	if p.state != Resolved {
		t.Errorf("Expected promise to be resolved; got %v", p.state)
	}

	if p.result != resolvedValue {
		t.Errorf("Expected promise result to be %v; got %v", resolvedValue, p.result)
	}
}

// 测试Promise被拒绝
func TestPromiseRejected(t *testing.T) {
	testError := errors.New("failure")
	p := NewPromise(func(resolve func(interface{}), reject func(error)) {
		reject(testError)
	})

	<-p.doneChan // 等待Promise拒绝
	if p.state != Rejected {
		t.Errorf("Expected promise to be rejected; got %v", p.state)
	}

	if p.err != testError {
		t.Errorf("Expected promise error to be %v; got %v", testError, p.err)
	}
}

// 测试Then链式调用
func TestPromiseThenChain(t *testing.T) {
	step1 := "step1"
	step2 := "step2"
	finalResult := "final result"
	p := NewPromise(func(resolve func(interface{}), reject func(error)) {
		resolve(step1)
	})

	p.Then(func(data interface{}) interface{} {
		if data != step1 {
			t.Errorf("Expected step1 data to be %v; got %v", step1, data)
		}
		return step2
	}).Then(func(data interface{}) interface{} {
		if data != step2 {
			t.Errorf("Expected step2 data to be %v; got %v", step2, data)
		}
		return finalResult
	}).Then(func(data interface{}) interface{} {
		if data != finalResult {
			t.Errorf("Expected final result to be %v; got %v", finalResult, data)
		}
		return nil
	})
}

// 测试Catch是否捕获错误
func TestPromiseCatch(t *testing.T) {
	testError := errors.New("failure")
	p := NewPromise(func(resolve func(interface{}), reject func(error)) {
		reject(testError)
	})

	caught := false
	p = p.Catch(func(err error) {
		caught = true
		if err != testError {
			t.Errorf("Expected caught error to be %v; got %v", testError, err)
		}
	})

	<-p.doneChan // 等待Catch执行
	if !caught {
		t.Error("Expected error to be caught")
	}
}

// 测试异步Promise解决
func TestPromiseAsyncResolve(t *testing.T) {
	resolvedValue := "async success"
	p := NewPromise(func(resolve func(interface{}), reject func(error)) {
		time.Sleep(1 * time.Second) // 模拟异步操作
		resolve(resolvedValue)
	})

	select {
	case <-p.doneChan: // 等待Promise解决
		if p.state != Resolved {
			t.Errorf("Expected promise to be resolved; got %v", p.state)
		}
		if p.result != resolvedValue {
			t.Errorf("Expected promise result to be %v; got %v", resolvedValue, p.result)
		}
	case <-time.After(2 * time.Second):
		t.Error("Promise resolution took too long")
	}
}
