package mixeddeadlock

import (
	"sync"
	"testing"
	"time"
)

func TestMixedDeadlock_GeneralCycle_3G_2Locks_2Chans(t *testing.T) {
	var l1, l2 sync.Mutex
	c1 := make(chan int) // unbuffered
	c2 := make(chan int)

	// step channels to enforce progress
	s12 := make(chan struct{})
	s23 := make(chan struct{})
	done := make(chan struct{})

	// g1: PCS on L1 then send on c1
	g1 := func() {
		l1.Lock()
		time.Sleep(5 * time.Millisecond)
		l1.Unlock()

		close(s12) // allow g2 to start receiving in CS
		c1 <- 1    // blocks until g2 receives
	}

	// g2: recv in CS on L1, then L1->L2, then PCS on L2 then send on c2
	g2 := func() {
		<-s12
		l1.Lock()
		v := <-c1 // recv inside CS on L1
		l2.Lock() // L1 -> L2 dependency
		l2.Unlock()
		l1.Unlock()

		l2.Lock()
		time.Sleep(5 * time.Millisecond)
		l2.Unlock()

		close(s23) // allow g3 to start receiving in CS
		c2 <- v    // blocks until g3 receives
	}

	// g3: recv in CS on L2, then L2->L1 and finish
	g3 := func() {
		<-s23
		l2.Lock()
		<-c2      // recv inside CS on L2
		l1.Lock() // L2 -> L1 dependency (closes cycle)
		l1.Unlock()
		l2.Unlock()
		close(done)
	}

	go g1()
	go g2()
	go g3()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("test timed out (deadlock)")
	}
}

func TestMixedDeadlock_GeneralCycle_CloseInChain(t *testing.T) {
	var l1, l2 sync.Mutex
	c1 := make(chan int) // unbuffered
	c2 := make(chan int) // closed later

	// g1: build first pair on L2 via MD2-3 on c1
	g1 := func() {
		l2.Lock()
		time.Sleep(10 * time.Millisecond)
		l2.Unlock() // PCS on L2

		c1 <- 1
	}

	// closer: PCS on L1, then close(c2) => MD2-3-Close with receiver-in-CS
	closer := func() {
		l1.Lock()
		time.Sleep(10 * time.Millisecond)
		l1.Unlock() // PCS on L1

		close(c2)
	}

	// g2: recv c1 inside CS on L2 (receiver side of first MD2-3),
	// then L2->L1, then recv-on-closed inside CS on L1 (receiver side of close MD2-3-Close),
	// then L1->L2 to close global cycle
	g2 := func() {
		l2.Lock()
		<-c1 // recv inside CS on L2
		// L2 -> L1
		l1.Lock()
		l1.Unlock()
		l2.Unlock()

		// wait so close definitely happened (not strictly required with unbuffered, but ok)
		time.Sleep(20 * time.Millisecond)

		l1.Lock()
		<-c2 // recv on closed inside CS on L1
		// L1 -> L2
		l2.Lock()
		l2.Unlock()
		l1.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); g1() }()
	go func() { defer wg.Done(); closer() }()
	go func() { defer wg.Done(); g2() }()
	wg.Wait()
}

func TestMixedDeadlock_GeneralCycle_WithSelect(t *testing.T) {
	var l1, l2 sync.Mutex
	c1 := make(chan int)
	c2 := make(chan int)

	s12 := make(chan struct{})
	s23 := make(chan struct{})
	done := make(chan struct{})

	g1 := func() {
		l1.Lock()
		time.Sleep(5 * time.Millisecond)
		l1.Unlock()

		close(s12)
		c1 <- 1
	}

	g2 := func() {
		<-s12
		l1.Lock()
		var v int
		select {
		case v = <-c1: // recv inside CS
		}
		l2.Lock() // L1 -> L2
		l2.Unlock()
		l1.Unlock()

		l2.Lock()
		time.Sleep(5 * time.Millisecond)
		l2.Unlock()

		close(s23)
		c2 <- v
	}

	g3 := func() {
		<-s23
		l2.Lock()
		select {
		case <-c2: // recv inside CS
		}
		l1.Lock() // L2 -> L1
		l1.Unlock()
		l2.Unlock()
		close(done)
	}

	go g1()
	go g2()
	go g3()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("test timed out (deadlock)")
	}
}
