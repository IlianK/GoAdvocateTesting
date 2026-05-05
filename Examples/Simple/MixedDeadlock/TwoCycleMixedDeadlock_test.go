// advocate/Examples/Examples_Simple/MixedDeadlock/mixed_deadlock_test.go

/*
------------------------------------------------------------
Mixed Deadlock Test (MDS-2)
------------------------------------------------------------

Each test corresponds to one theoretical case:
- MD2-1 : both inside CS  (symmetric)
- MD2-2 : sender inside, receiver after CS (lock→channel)
- MD2-3 : sender after CS, receiver inside (channel→lock)

Close-Recv variants: MD2-X-Close

Variants:
- U: unbuffered channel
- B: buffered channel

LockType:
- READ/READ
- READ/WRITE
- WRITE/WRITE
*/

package mixeddeadlock

import (
	"sync"
	"testing"
	"time"
)

// ------------------------------------------------------------
// MD2-1: Both sender and receiver in CS
// ------------------------------------------------------------

func TestMixedDeadlock_TEST(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	go func() {
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()
	}()

	time.Sleep(50 * time.Millisecond) // let sender go first
	m.Lock()
	<-c // receive inside CS
	m.Unlock()
}

// MD2-1B: Buffered Variant
func TestMixedDeadlock_MD2_1B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	sender := func() {
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()

	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let sender go first
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}

// MD2-1U: Unbuffered Close Variant
func TestMixedDeadlock_MD2_1_CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // buffered

	closer := func() {
		m.Lock()
		close(c) // close inside CS
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let closer go first
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(closer, receiver)
}

// MD2-1B: Buffered Close Variant
func TestMixedDeadlock_MD2_1_CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	closer := func() {
		m.Lock()
		close(c) // close inside CS
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let closer go first
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(closer, receiver)
}

// MD2-1B: Buffered READ/WRTIE Variant
func TestMixedDeadlock_MD_2_1B_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1)

	writer := func() {
		rw.Lock()
		c <- 1 // send inside CS
		rw.Unlock()
	}

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender finish PCS
		rw.RLock()
		<-c // receive in CS
		rw.RUnlock()
	}

	run2(reader, writer)
}

// ------------------------------------------------------------
// MD2-2: Sender inside CS, Receiver with PCS
// ------------------------------------------------------------

// MD2-2U: Unbuffered Variant
func TestMixedDeadlock_MD2_2U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		time.Sleep(50 * time.Millisecond) // let receiver complete PCS
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

// MD2-2U: Unbuffered READ/WRTIE Variant
func TestMixedDeadlock_MD_2_2U_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int)

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
// MD2-3: Sender with PCS, Receiver inside CS
// ------------------------------------------------------------

// MD2-3U: Unbuffered Variant
func TestMixedDeadlock_MD2_3U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		c <- 1 // send after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) //let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}

// MD2-3B: Buffered Variant
func TestMixedDeadlock_MD2_3B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1 // send after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}

// MD2-3U: Unbuffered Close Variant
func TestMixedDeadlock_MD_2_3CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	closer := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		close(c) // close after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let closer complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(receiver, closer)
}

// MD2-3B: Buffered Close Variant
func TestMixedDeadlock_MD_2_3CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	closer := func() {
		m.Lock()
		time.Sleep(50 * time.Millisecond)
		m.Unlock()
		close(c) // close after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // let closer complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(receiver, closer)
}

// MD2-3U: Unbuffered READ/WRTIE Variant
func TestMixedDeadlock_MD_2_3U_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int)

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender finish PCS
		rw.RLock()
		<-c // receive inside CS
		rw.RUnlock()
	}

	writer := func() {
		rw.Lock()
		time.Sleep(10 * time.Millisecond)
		rw.Unlock() // PCS
		c <- 1      // send after PCS
	}

	run2(reader, writer)
}

// MD2-3B: Buffered READ/WRTIE Variant
func TestMixedDeadlock_MD_2_3B_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1)

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender finish PCS
		rw.RLock()
		<-c // receive inside CS
		rw.RUnlock()
	}

	writer := func() {
		rw.Lock()
		time.Sleep(10 * time.Millisecond)
		rw.Unlock() // PCS
		c <- 1      // send after PCS
	}

	run2(reader, writer)
}
