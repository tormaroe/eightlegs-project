package queue

import "sync"

type atomicCount struct {
	mut *sync.Mutex
	v   int
}

func (a *atomicCount) inc() {
	a.mut.Lock()
	defer a.mut.Unlock()
	a.v++
}

func (a *atomicCount) dec() {
	a.mut.Lock()
	defer a.mut.Unlock()
	a.v--
}

func (a *atomicCount) val() int {
	return a.v
}
