package deadlocks

import (
	"sync"
	"testing"
	"time"
)

func Test_CyclicDeadlock_ABBA_TwoMutexes(t *testing.T) {
	var a, b sync.Mutex
	start := make(chan struct{})

	go func() { // G1
		<-start
		a.Lock()
		time.Sleep(1 * time.Millisecond)
		b.Lock()
	}()

	go func() { // G2
		<-start
		b.Lock()
		time.Sleep(1 * time.Millisecond)
		a.Lock()
	}()

	close(start)
	time.Sleep(50 * time.Millisecond)
}

func Test_CyclicDeadlock_ThreeWayMutexCycle(t *testing.T) {
	var a, b, c sync.Mutex
	start := make(chan struct{})

	go func() { // G1: A then B
		<-start
		a.Lock()
		time.Sleep(1 * time.Millisecond)
		b.Lock()
	}()

	go func() { // G2: B then C
		<-start
		b.Lock()
		time.Sleep(1 * time.Millisecond)
		c.Lock()
	}()

	go func() { // G3: C then A
		<-start
		c.Lock()
		time.Sleep(1 * time.Millisecond)
		a.Lock()
	}()

	close(start)
	time.Sleep(50 * time.Millisecond)
}

func Test_CyclicDeadlock_RWMutex_WriterThenReader(t *testing.T) {
	var rw sync.RWMutex
	start := make(chan struct{})

	// G1 holds RLock and then tries to take Lock => blocks.
	go func() {
		<-start
		rw.RLock()
		time.Sleep(2 * time.Millisecond)
		rw.Lock() // blocks
	}()

	// G2 tries to take Lock while reader holds RLock => blocks too.
	go func() {
		<-start
		time.Sleep(1 * time.Millisecond)
		rw.Lock() // blocks because of G1 RLock
	}()

	close(start)
	time.Sleep(50 * time.Millisecond)
}

func Test_CyclicDeadlock_ChannelAndMutex(t *testing.T) {
	var mu sync.Mutex
	ch := make(chan struct{})
	start := make(chan struct{})

	go func() { // G1
		<-start
		mu.Lock()
		<-ch // waits for send while holding mu
		mu.Unlock()
	}()

	go func() { // G2
		<-start
		time.Sleep(1 * time.Millisecond)
		mu.Lock()        // blocks (mu held by G1)
		ch <- struct{}{} // can't be reached
		mu.Unlock()
	}()

	close(start)
	time.Sleep(50 * time.Millisecond)
}
