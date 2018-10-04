package mgo

// 单线程环境下会引发性能降低
// runtime.NumCPU() == 1 || runtime.GOMAXPROCS == 1

import (
	"runtime"
	"sync/atomic"
)

type SpinLock uint32

// 不同于内核，内核会关闭抢占，等待过程不会休眠，所以持锁后也不能被抢占和休眠
// 如果在内核中断上下文有使用，还必须屏蔽中断，避免持锁过程被中断导致中断上下文发生死锁
func (self *SpinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(self), 0, 1) {
		runtime.Gosched()
	}
}

func (self *SpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(self), 0)
}
