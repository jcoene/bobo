package bobo

import (
	"fmt"
	"time"
)

type result struct {
	obj   interface{}
	found bool
	err   error
}

type TimeoutError struct {
	Duration time.Duration
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("timed out after %v", e.Duration)
}

func Timeout(dur time.Duration, fn func() (interface{}, bool, error)) (interface{}, bool, error) {
	ch := make(chan result, 1)
	go func() {
		obj, found, err := fn()
		ch <- result{obj, found, err}
	}()

	select {
	case res := <-ch:
		return res.obj, res.found, res.err
	case <-time.After(dur):
		return nil, false, TimeoutError{dur}
	}
}
