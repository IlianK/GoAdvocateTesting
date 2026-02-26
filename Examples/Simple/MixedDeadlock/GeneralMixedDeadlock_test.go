package mixeddeadlock

import (
	"sync"
	"testing"
	"time"
)

/*
GeneralCycle MD tests

Goal: 3 distinct “general cycle” mixed-deadlock patterns (not redundant),
while keeping them trace-friendly and deterministic.

Conventions:
- We use unbuffered channels to force concrete send/recv rendezvous pairs.
- We enforce ordering via step channels (s12, s23, …).
- We *finish* the test (done closes) so traces are clean; the analyzer should still
  see the mixed cycle in the offline graph.
*/

// -----------------------------------------------------------------------------
// 1) GeneralCycle: 3 goroutines, 2 locks, 2 channels (no select)
// Pattern intent:
//   - Pair edges from c1 and c2 rendezvous
//   - Lock edges via L1->L2 and L2->L1 cross acquisition
//
// Distinct from the “with select” version.
// -----------------------------------------------------------------------------
func TestMixedDeadlock_GeneralCycle_3G_2Locks_2Chans(t *testing.T) {
	var l1, l2 sync.Mutex
	c1 := make(chan int) // unbuffered rendezvous
	c2 := make(chan int) // unbuffered rendezvous

	// Step channels to enforce a deterministic schedule.
	s12 := make(chan struct{})
	s23 := make(chan struct{})
	done := make(chan struct{})

	// G1: PCS on L1, then send on c1
	go func() {
		l1.Lock()
		time.Sleep(5 * time.Millisecond)
		l1.Unlock() // PCS(L1)

		close(s12)
		c1 <- 1 // rendezvous with G2
	}()

	// G2: recv c1 while holding L1 (recv-in-CS on L1),
	//     then create L1->L2 dependency, then PCS(L2) and send on c2
	go func() {
		<-s12
		l1.Lock()
		v := <-c1 // recv inside CS(L1)

		l2.Lock() // L1 -> L2 edge
		l2.Unlock()
		l1.Unlock()

		l2.Lock()
		time.Sleep(5 * time.Millisecond)
		l2.Unlock() // PCS(L2)

		close(s23)
		c2 <- v // rendezvous with G3
	}()

	// G3: recv c2 while holding L2 (recv-in-CS on L2),
	//     then create L2->L1 dependency to close the cycle
	go func() {
		<-s23
		l2.Lock()
		<-c2 // recv inside CS(L2)

		l1.Lock() // L2 -> L1 edge (closes cycle)
		l1.Unlock()
		l2.Unlock()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("test timed out (expected progress but got stuck)")
	}
}

// -----------------------------------------------------------------------------
// 2) GeneralCycle: 3 goroutines, 2 locks, 2 channels + select receives
// Pattern intent:
//   - Same *high-level* resources (L1/L2 + c1/c2) but introduces Select events
//     (SS) in the trace, which is useful to validate select-aware pairing paths.
//
// Distinct from #1 because the receives are via `select`.
// -----------------------------------------------------------------------------
func TestMixedDeadlock_GeneralCycle_3G_2Locks_2Chans_WithSelectRecv(t *testing.T) {
	var l1, l2 sync.Mutex
	c1 := make(chan int)
	c2 := make(chan int)

	s12 := make(chan struct{})
	s23 := make(chan struct{})
	done := make(chan struct{})

	// G1: PCS(L1) then send on c1
	go func() {
		l1.Lock()
		time.Sleep(5 * time.Millisecond)
		l1.Unlock() // PCS(L1)

		close(s12)
		c1 <- 1
	}()

	// G2: recv c1 via select while holding L1; create L1->L2; PCS(L2); send c2
	go func() {
		<-s12
		l1.Lock()
		var v int
		select {
		case v = <-c1: // recv inside CS(L1), but via select (SS)
		}

		l2.Lock() // L1 -> L2
		l2.Unlock()
		l1.Unlock()

		l2.Lock()
		time.Sleep(5 * time.Millisecond)
		l2.Unlock() // PCS(L2)

		close(s23)
		c2 <- v
	}()

	// G3: recv c2 via select while holding L2; create L2->L1; done
	go func() {
		<-s23
		l2.Lock()
		select {
		case <-c2: // recv inside CS(L2), via select (SS)
		}

		l1.Lock() // L2 -> L1 (closes cycle)
		l1.Unlock()
		l2.Unlock()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("test timed out (expected progress but got stuck)")
	}
}

// -----------------------------------------------------------------------------
// 3) GeneralCycle: 4 goroutines, 2 locks, 3 channels (longer cycle)
// Pattern intent:
//   - Not just a stylistic variant: adds an extra channel/pair node,
//     making the mixed cycle “longer” in terms of (Lock + Pair) alternation.
//
// Distinct from #1/#2 because it uses an additional rendezvous hop.
// -----------------------------------------------------------------------------
func TestMixedDeadlock_GeneralCycle_4G_2Locks_3Chans(t *testing.T) {
	var l1, l2 sync.Mutex
	c12 := make(chan int) // G1 -> G2
	c23 := make(chan int) // G2 -> G3
	c34 := make(chan int) // G3 -> G4

	s12 := make(chan struct{})
	s23 := make(chan struct{})
	s34 := make(chan struct{})
	done := make(chan struct{})

	// G1: PCS(L1) then send on c12
	go func() {
		l1.Lock()
		time.Sleep(5 * time.Millisecond)
		l1.Unlock() // PCS(L1)

		close(s12)
		c12 <- 1
	}()

	// G2: recv c12 in CS(L1), create L1->L2, then send on c23
	go func() {
		<-s12
		l1.Lock()
		v := <-c12 // recv in CS(L1)

		l2.Lock() // L1 -> L2
		l2.Unlock()
		l1.Unlock()

		close(s23)
		c23 <- v
	}()

	// G3: PCS(L2) then recv c23 outside CS, then send c34 in CS(L2)
	// This introduces a PCS-style edge and a send-in-CS edge around the same lock.
	go func() {
		<-s23
		l2.Lock()
		time.Sleep(5 * time.Millisecond)
		l2.Unlock() // PCS(L2)

		v := <-c23 // recv after PCS(L2) (outside CS)

		close(s34)
		l2.Lock()
		c34 <- v // send in CS(L2)
		l2.Unlock()
	}()

	// G4: recv c34 (outside CS), then create L2->L1 to close the cycle
	go func() {
		<-s34
		<-c34

		l2.Lock()
		l1.Lock() // L2 -> L1 (closes cycle)
		l1.Unlock()
		l2.Unlock()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("test timed out (expected progress but got stuck)")
	}
}
