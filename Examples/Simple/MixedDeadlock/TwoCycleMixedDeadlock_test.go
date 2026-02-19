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
	"advocate"
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
// MD2-1: Both sender and receiver in CS
// ------------------------------------------------------------

/*
// MD2-U: Buffered Variant
func TestMixedDeadlock_MD2_1U((t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	sender := func() {
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()

	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}
*/

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

// MD2-2B: Buffered Variant
func TestMixedDeadlock_MD2_2B(t *testing.T) {
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

// ------------------------------------------------------------
// CLOSE TESTS: MDX-Y-CLOSE VARIANTS
// ------------------------------------------------------------

// MD2-1U: Unbuffered Variant
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

// MD2-1B: Buffered Variant
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

// MD-CloseU: Unbuffered Variant
func TestMixedDeadlock_MD_2_2CloseU(t *testing.T) {
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
func TestMixedDeadlock_MD_2_2CloseB(t *testing.T) {
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

// MD-CloseU: Unbuffered Variant
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

// MD-CloseB: Buffered Variant (Mirror of MD-2-3B)
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

// ------------------------------------------------------------
// LOCKTYPE TESTS: READ/WRITE MD-Cases
// ------------------------------------------------------------

// READ/WRTIE MD2-1B
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

// READ/WRTIE MD2-2U
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

// READ/WRTIE MD2-2B
func TestMixedDeadlock_MD_2_2B_RW_FP(t *testing.T) {
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

// READ/WRTIE MD2-3U
func TestMixedDeadlock_MD_2_3_RW(t *testing.T) {
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

// READ/WRTIE MD2-3B
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
// MORE THAN ONE CS
// ------------------------------------------------------------

func TestMixedDeadlock_DoubleCS_Send(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		c <- 1
		m.Unlock()
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
		m.Lock()
		c <- 1
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
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_DoubleCS_Recv_2(t *testing.T) {
	// ======= Preamble Start =======
  advocate.InitReplay("rewrittenTrace_1", 5, true)
  defer advocate.FinishReplay()
  // ======= Preamble End =======
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
		m.Lock()
		<-c
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_DoubleCS_Send_Recv_1(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		m.Lock()
		c <- 1
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		m.Lock()
		<-c
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
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c
		m.Unlock()
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

// ------------------------------------------------------------
// FALSE POSITIVE TESTS
// ------------------------------------------------------------

// READ/READ
func TestMixedDeadlock_MD_Read(t *testing.T) {
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

// No Mixed Deadlock: both outside CS
func TestMixedDeadlock_No_MD_BeforeCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		c <- 1 // before CS
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	receiver := func() {
		<-c // before CS
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}

// No Mixed Deadlock: Both after CS
func TestMixedDeadlock_No_MD_AfterPCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1 // after PCS
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // after PCS
	}

	run2(sender, receiver)
}

// No Mixed Deadlock: Different Locks
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

// WMHB: No Mixed Deadlock due to Fork Must-Order
func TestMixedDeadlock_No_MD_ForkMustOrder(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)
	done := make(chan struct{})

	// Parent acquires BEFORE forking
	m.Lock()

	// Child after
	go func() {
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
		close(done)
	}()

	// Parent sends while STILL in CS (would look like MD2-1 if reordering was allowed),
	// but acquires are must-ordered by the fork, so the detector should NOT flag it.

	c <- 1 // send inside CS, non-blocking due to buffer

	m.Unlock()
	<-done
}
