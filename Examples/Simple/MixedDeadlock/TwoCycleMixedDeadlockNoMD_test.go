package mixeddeadlock

import (
	"sync"
	"testing"
	"time"
)

// ------------------------------------------------------------
// No_MD 2-2 Cases
// ------------------------------------------------------------

// MD2-2B: Buffered Variant
func TestMixedDeadlock_No_MD_2_2B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		// no sleep, always non-blocking due to buffer
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		<-c // receive after PCS
	}

	run2(sender, receiver)
}

// MD-CloseU: Unbuffered Variant
func TestMixedDeadlock_No_MD_2_2CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	closer := func() {
		// no sleep, recv is blocking until close
		m.Lock()
		close(c) // close in CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		<-c // recv with PCS
	}

	run2(receiver, closer)
}

// MD-CloseB: Buffered Variant (Mirror of MD-2-2B)
func TestMixedDeadlock_No_MD_2_2CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // unbuffered

	closer := func() {
		// no sleep, recv is blocking until close
		m.Lock()
		close(c) // close in CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		<-c // recv with PCS
	}

	run2(receiver, closer)

}

// READ/WRTIE MD2-2B
func TestMixedDeadlock_No_MD_2_2B_RW_FP(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1)

	writer := func() {
		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
		rw.Lock()
		c <- 1 // send inside CS
		rw.Unlock()
	}

	reader := func() {
		rw.RLock()
		time.Sleep(10 * time.Millisecond)
		rw.RUnlock() // PCS
		<-c          // receive after PCS
	}

	run2(reader, writer)
}

// ------------------------------------------------------------
// No_MD Other Cases
// ------------------------------------------------------------

// READ/READ
func TestMixedDeadlock_No_MD_Read(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int)

	reader_1 := func() {
		rw.RLock()
		<-c
		rw.RUnlock()
	}

	reader_2 := func() {
		rw.RLock()
		c <- 1
		rw.RUnlock()
	}

	run2(reader_1, reader_2)
}

// Both before CS
func TestMixedDeadlock_No_MD_BeforeCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		c <- 1
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	receiver := func() {
		<-c
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}

// Both after CS
func TestMixedDeadlock_No_MD_AfterPCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c
	}

	run2(sender, receiver)
}

// Different locks
func TestMixedDeadlock_No_MD_DifferentLocks(t *testing.T) {
	var m1, m2 sync.Mutex
	c := make(chan int)

	sender := func() {
		m1.Lock()
		c <- 1
		m1.Unlock()
	}

	receiver := func() {
		m2.Lock()
		<-c
		m2.Unlock()
	}

	run2(sender, receiver)
}

// WMHB Fork ordering
func TestMixedDeadlock_No_MD_ForkMustOrder(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)
	done := make(chan struct{})

	m.Lock()

	go func() {
		m.Lock()
		<-c
		m.Unlock()
		close(done)
	}()

	c <- 1 // send inside CS (buffered)
	m.Unlock()
	<-done
}
