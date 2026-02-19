package deadlocks

import (
	"sync"
	"testing"
	"time"
)

// 1) Two‐lock inversion (classic two‐goroutine cycle)
func TestCycleTwoLocks(t *testing.T) {
	var a, b sync.Mutex

	go func() {
		a.Lock()
		time.Sleep(10 * time.Millisecond)
		b.Lock()
		b.Unlock()
		a.Unlock()
	}()

	b.Lock()
	time.Sleep(5 * time.Millisecond)
	a.Lock()
	a.Unlock()
	b.Unlock()
}

// 2) Three‐lock cycle (A→B→C→A)
func TestCycleThreeLocks(t *testing.T) {
	var x, y, z sync.Mutex

	go func() {
		x.Lock()
		time.Sleep(5 * time.Millisecond)
		y.Lock()
		y.Unlock()
		x.Unlock()
	}()
	go func() {
		y.Lock()
		time.Sleep(5 * time.Millisecond)
		z.Lock()
		z.Unlock()
		y.Unlock()
	}()
	// main goroutine forms the third link:
	z.Lock()
	time.Sleep(5 * time.Millisecond)
	x.Lock() // deadlock on x
	x.Unlock()
	z.Unlock()
}


// 3) Cond deadlock: Wait with no Signal
func TestCondDeadlock(t *testing.T) {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	go func() {
		time.Sleep(10 * time.Millisecond)
		// no cond.Signal()
	}()
	mu.Lock()
	cond.Wait() // blocks forever
	mu.Unlock()
}
