package executor

import "sync"

var (
	once sync.Once
	inst *Executor
)

type Executor struct {
}

func Get() *Executor {
	once.Do(func() {
		inst = &Executor{}
	})
	return inst
}
