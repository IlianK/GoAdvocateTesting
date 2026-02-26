package results_scenarios

import (
	"context"
	"sync"
	"testing"
	"time"
)

// A01: Actual Send on Closed Channel
func TestA01_SendOnClosedChannel(t *testing.T) {
	c := make(chan int)
	close(c)
	defer func() { _ = recover() }() // prevent panic from failing the entire suite
	c <- 1
}

// A02: Actual Receive on Closed Channel
func TestA02_RecvOnClosedChannel(t *testing.T) {
	c := make(chan int)
	close(c)
	_ = <-c // receive succeeds with zero value, but is "recv on closed"
}

// A03: Actual Close on Closed Channel
func TestA03_CloseOnClosedChannel(t *testing.T) {
	c := make(chan int)
	close(c)
	defer func() { _ = recover() }()
	close(c)
}

// A04: Actual Close on nil channel
func TestA04_CloseOnNilChannel(t *testing.T) {
	var c chan int
	defer func() { _ = recover() }()
	close(c)
}

// A05: Actual negative WaitGroup
func TestA05_NegativeWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
	defer func() { _ = recover() }() // wg counter going negative panics
	wg.Done()
}

// A06: Actual unlock of not locked mutex
func TestA06_UnlockNotLockedMutex(t *testing.T) {
	var m sync.Mutex
	defer func() { _ = recover() }() // unlocking unlocked mutex panics
	m.Unlock()
}

// A07: Actual Leak (Non-Cyclic blocking bug)  (goroutine blocks forever)
func TestA07_ActualLeak_UnbufferedRecvNoPartner(t *testing.T) {
	c := make(chan struct{})
	go func() {
		<-c // blocks forever
	}()
	time.Sleep(10 * time.Millisecond) // keep trace alive briefly
}

// A08: Actual Deadlock (Cyclic blocking bug)  (classic two-mutex cycle)
func TestA08_ActualDeadlock_CyclicMutex(t *testing.T) {
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

// A09: Concurrent Receives on the same channel
func TestA09_ConcurrentReceivesSameChannel(t *testing.T) {
	c := make(chan int)

	ready := make(chan struct{}, 2)
	go func() { ready <- struct{}{}; <-c }()
	go func() { ready <- struct{}{}; <-c }()
	<-ready
	<-ready

	// One send; one recv matches, the other remains blocked -> concurrent recv observed
	c <- 1
	time.Sleep(10 * time.Millisecond)
}

// P01: Possible Send on Closed Channel (send races with close; in this run, send happens first)
func TestP01_PossibleSendOnClosed(t *testing.T) {
	c := make(chan int, 1)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { // sender
		defer wg.Done()
		<-start
		c <- 1 // in this schedule, send happens before close
	}()

	go func() { // closer
		defer wg.Done()
		<-start
		time.Sleep(1 * time.Millisecond)
		close(c)
	}()

	close(start)
	wg.Wait()
}

// P02: Possible Receive on Closed Channel (recv races with close; in this run, recv happens first)
func TestP02_PossibleRecvOnClosed(t *testing.T) {
	c := make(chan int, 1)
	c <- 1

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { // receiver
		defer wg.Done()
		<-start
		<-c // in this schedule, receive consumes before close
	}()

	go func() { // closer
		defer wg.Done()
		<-start
		time.Sleep(1 * time.Millisecond)
		close(c)
	}()

	close(start)
	wg.Wait()
}

// P03: Possible Negative WaitGroup Counter (adds/dones concurrent; schedule avoids actual negative)
func TestP03_PossibleNegativeWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	var done sync.WaitGroup
	done.Add(3)

	go func() { // Add(1)
		defer done.Done()
		<-start
		wg.Add(1)
	}()

	go func() { // Add(1)
		defer done.Done()
		<-start
		wg.Add(1)
	}()

	go func() { // Done twice (could go negative if Add doesn't happen)
		defer done.Done()
		<-start
		time.Sleep(2 * time.Millisecond) // bias toward Add first so it's only "possible"
		wg.Done()
		wg.Done()
	}()

	close(start)
	done.Wait()
}

// P04: Possible unlock of not locked mutex (unlock may race with lock; schedule makes it not panic)
func TestP04_PossibleUnlockNotLockedMutex(t *testing.T) {
	var m sync.Mutex
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { // locker
		defer wg.Done()
		<-start
		m.Lock()
		time.Sleep(2 * time.Millisecond)
		m.Unlock()
	}()

	go func() { // unlocker (could be too early => actual panic, but we delay)
		defer wg.Done()
		<-start
		time.Sleep(1 * time.Millisecond)
		m.Unlock() // in this schedule, should unlock after lock acquired -> no panic
	}()

	close(start)
	wg.Wait()
}

// P05: Possible cyclic deadlock (two-lock ordering inversion but schedule avoids deadlock)
func TestP05_PossibleCyclicDeadlock(t *testing.T) {
	var a, b sync.Mutex
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { // takes a then b
		defer wg.Done()
		<-start
		a.Lock()
		time.Sleep(1 * time.Millisecond)
		b.Lock()
		b.Unlock()
		a.Unlock()
	}()

	go func() { // takes b then a, but later (avoid actual deadlock)
		defer wg.Done()
		<-start
		time.Sleep(3 * time.Millisecond)
		b.Lock()
		a.Lock()
		a.Unlock()
		b.Unlock()
	}()

	close(start)
	wg.Wait()
}

// L00: Leak (generic) — use a goroutine blocked on select with no cases firing
func TestL00_Leak_Generic(t *testing.T) {
	ch := make(chan struct{})
	go func() {
		select {
		case <-ch:
		}
	}()
	time.Sleep(10 * time.Millisecond)
}

// L01: Leak on unbuffered channel with possible partner
func TestL01_LeakUnbuffered_WithPartner(t *testing.T) {
	c := make(chan int)

	go func() { c <- 1 }() // send #1
	go func() { <-c }()    // recv #1 partners with send #1

	go func() { c <- 2 }() // send #2 leaks, but there *exists* a recv op on c (possible partner)
	time.Sleep(10 * time.Millisecond)
}

// L02: Leak on unbuffered channel without possible partner
func TestL02_LeakUnbuffered_WithoutPartner(t *testing.T) {
	c := make(chan int)
	go func() { <-c }() // recv leaks; no sender exists
	time.Sleep(10 * time.Millisecond)
}

// L03: Leak on buffered channel with possible partner
func TestL03_LeakBuffered_WithPartner(t *testing.T) {
	c := make(chan int, 1)

	go func() { c <- 1 }() // fills buffer
	go func() { <-c }()    // possible partner exists

	go func() { c <- 2 }() // blocks if buffer is full at that moment => leak with partner
	time.Sleep(10 * time.Millisecond)
}

// L04: Leak on buffered channel without possible partner
func TestL04_LeakBuffered_WithoutPartner(t *testing.T) {
	c := make(chan int, 1)
	go func() { <-c }() // recv on empty buffered channel; no sender => leak
	time.Sleep(10 * time.Millisecond)
}

// L05: Leak on nil channel
func TestL05_LeakNilChannel(t *testing.T) {
	var c chan int         // nil
	go func() { c <- 1 }() // blocks forever
	time.Sleep(10 * time.Millisecond)
}

// L06: Leak on select with possible partner
func TestL06_LeakSelect_WithPartner(t *testing.T) {
	c := make(chan int)

	go func() { <-c }() // possible partner exists

	go func() {
		select {
		case c <- 1: // leaks because no recv scheduled at same time (or already consumed)
		}
	}()
	time.Sleep(10 * time.Millisecond)
}

// L07: Leak on select without possible partner
func TestL07_LeakSelect_WithoutPartner(t *testing.T) {
	c := make(chan int)
	go func() {
		select {
		case c <- 1: // no receiver exists => leak
		}
	}()
	time.Sleep(10 * time.Millisecond)
}

// L08: Leak on sync.Mutex (lock attempt blocks; last lock exists)
func TestL08_LeakMutex(t *testing.T) {
	var m sync.Mutex
	m.Lock() // last lock
	go func() {
		m.Lock() // leaks here
	}()
	time.Sleep(10 * time.Millisecond)
}

// L09: Leak on sync.WaitGroup (Wait blocks forever)
func TestL09_LeakWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Wait() // leaks forever because Done never called
	}()
	time.Sleep(10 * time.Millisecond)
}

// L10: Leak on sync.Cond (Wait without signal/broadcast)
func TestL10_LeakCond(t *testing.T) {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	go func() {
		cond.L.Lock()
		cond.Wait() // leaks forever
		cond.L.Unlock()
	}()
	time.Sleep(10 * time.Millisecond)
}

// L11: Leak on channel or select on context (block on ctx.Done that never closes)
func TestL11_LeakContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel // never called

	go func() {
		<-ctx.Done() // leaks forever
	}()
	time.Sleep(10 * time.Millisecond)
}
