package mixeddeadlock

import (
	"sync"
	"testing"
	"time"
)

// Helper
func run2(a, b func()) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); a() }()
	go func() { defer wg.Done(); b() }()
	wg.Wait()
}

// ------------------------------------------------------------
// RWMutex Variants
// ------------------------------------------------------------

// Multiple Read Locks/Unlocks
func TestRWMutex_DoubleRLockCounter_MD2_1B(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1) // buffered für MD2-1

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender go first
		rw.RLock()
		rw.RLock()
		rw.RUnlock() // Reader Count: 2 -> 1
		time.Sleep(30 * time.Millisecond)
		// Receive still in CS
		<-c

		rw.RUnlock()
	}

	sender := func() {
		rw.Lock()
		c <- 1 // Send in CS
		rw.Unlock()
	}

	run2(reader, sender)
}

// Multiple Read Locks/Unlocks with Unbuffered Channel
func TestRWMutex_DoubleRUnlock_MD2_2U(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int) // buffered

	reader := func() {
		rw.RLock()
		rw.RLock()
		rw.RUnlock()
		rw.RUnlock()
		<-c // after PCS
	}

	sender := func() {
		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
		rw.Lock()
		c <- 1 // in CS
		rw.Unlock()
	}

	run2(reader, sender)
}

// ------------------------------------------------------------
// MULTIPLE CS
// ------------------------------------------------------------

func TestMixedDeadlock_DoubleCS_Send(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()

	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_DoubleCS_Send_2(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		c <- 1
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(150 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_DoubleCS_Recv(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_DoubleCS_Recv_2(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		<-c
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_DoubleCS_Send_Recv_1(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond) // CS1
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		c <- 1 // CS2 with send
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		time.Sleep(10 * time.Millisecond) // CS1
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		<-c // CS2 with receive
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_DoubleCS_Send_Recv_2(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()

		time.Sleep(10 * time.Millisecond)

		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}

/*
Cyclic dependency between D1 and D3.

There are no further cyclic dependencies (pl check).

This shows that if we only record the "last" dependency per thread,
we may run into false negatives.
*/
func TestMixedDeadlock_MultiDep_LastOnly(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered
	s := make(chan int, 1) // buffered

	receiver := func() { // G2
		// D3 = (G2, rcv(c), {m})
		m.Lock()
		<-c
		m.Unlock()

		s <- 1 // allow G1 to proceed
		<-s    // handshake

		// D4 = (G2, rcv(c), {m})
		m.Lock()
		<-c
		m.Unlock()
	}

	sender := func() { // G1
		// D1 = (G1, snd(c), {m})
		m.Lock()
		c <- 1
		m.Unlock()

		<-s // sync with G2

		// D2 = (G1, snd(c), {m})
		m.Lock()
		s <- 1
		c <- 1
		m.Unlock()
	}

	run2(receiver, sender)
}

/*
Cyclic dependency between D2 and D3, and between D2 and D4.
But D1 is not involved in any cyclic dependency.

This shows that if we only record the "first" dependency per thread,
we may run into false negatives.
*/
func TestMixedDeadlock_MultiDep_FirstOnly(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered
	s := make(chan int, 1) // buffered

	receiver := func() { // G2
		s <- 1 // let G1 continue

		// D3 = (G2, rcv(c), {m})
		m.Lock()
		<-c
		m.Unlock()

		time.Sleep(20 * time.Millisecond) // remove => likely deadlock

		// D4 = (G2, rcv(c), {m})
		m.Lock()
		<-c
		m.Unlock()
	}

	sender := func() { // G1
		// D1 = (G1, snd(c), {m})
		m.Lock()
		c <- 1
		m.Unlock()

		<-s // sync with G2

		// D2 = (G1, snd(c), {m})
		m.Lock()
		s <- 1
		c <- 1
		m.Unlock()
	}

	run2(receiver, sender)
}

func TestMixedDeadlock_MD3_ThreeRoutine(t *testing.T) {
	var x, y sync.Mutex
	c := make(chan int, 1) // buffered

	// R2: hold x, send on c (CS on x), release x
	r2 := func() {
		x.Lock()
		c <- 1 // send while holding x CD_R2
		x.Unlock()
	}

	// R3: hold x, receive on c (CS on x), then acquire y
	// The sleep ensures R2 sends first in the working trace
	r3 := func() {
		time.Sleep(60 * time.Millisecond) // let R2 go first
		x.Lock()
		<-c // receive while holding x CD_R3
		x.Unlock()

		y.Lock() // RD_R3: acquire y after releasing x
		y.Unlock()
	}

	// R4: hold y first, then acquire x
	// ensures R4 does not interfere with R2/R3 in the working trace.
	r4 := func() {
		time.Sleep(120 * time.Millisecond) // let R2 and R3 finish first
		y.Lock()                           // hold y RD_R4 lockset = {y}
		x.Lock()                           // want x RD_R4 lock = x
		x.Unlock()
		y.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); r2() }()
	go func() { defer wg.Done(); r3() }()
	go func() { defer wg.Done(); r4() }()
	wg.Wait()
}
