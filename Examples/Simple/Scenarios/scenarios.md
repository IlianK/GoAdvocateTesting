# Results Scenarios

This folder contains **minimal, focused test cases** that exercise the analyzer’s **result categories** in both the *Actual* and *Possible* families, as well as the various *Leak* classifications.  
Each test is designed to trigger **exactly one primary scenario** (or as close as possible), while keeping the run **trace-friendly** (tests stay alive briefly so the trace can capture the relevant operations).

## Actual Misuse / Actual Bugs (`Axx`)


### `TestA01_SendOnClosedChannel`
- **Scenario:** Close a channel, then perform a send on it.
- **Expected outcome:** **Actual send on closed channel** (panic in Go; recovered in test) — `A01`.

```go
// A01: Actual Send on Closed Channel
func TestA01_SendOnClosedChannel(t *testing.T) {
	c := make(chan int)
	close(c)
	defer func() { _ = recover() }() // prevent panic from failing the entire suite
	c <- 1
}
```


### `TestA02_RecvOnClosedChannel`
- **Scenario:** Close a channel, then receive from it.
- **Expected outcome:** **Actual receive on closed channel** — `A02`.

```go
// A02: Actual Receive on Closed Channel
func TestA02_RecvOnClosedChannel(t *testing.T) {
	c := make(chan int)
	close(c)
	_ = <-c // receive succeeds with zero value, but is "recv on closed"
}
```


### `TestA03_CloseOnClosedChannel`
- **Scenario:** Close a channel twice.
- **Expected outcome:** **Actual close on closed channel** (panic in Go; recovered in test) — `A03`.

```go
// A03: Actual Close on Closed Channel
func TestA03_CloseOnClosedChannel(t *testing.T) {
	c := make(chan int)
	close(c)
	defer func() { _ = recover() }()
	close(c)
}
```


### `TestA04_CloseOnNilChannel`
- **Scenario:** Close a nil channel.
- **Expected outcome:** **Actual close on nil channel** (panic in Go; recovered in test) — `A04`.

```go
// A04: Actual Close on nil channel
func TestA04_CloseOnNilChannel(t *testing.T) {
	var c chan int
	defer func() { _ = recover() }()
	close(c)
}
```


### `TestA05_NegativeWaitGroup`
- **Scenario:** Call `Done()` more often than `Add()` (counter goes negative).
- **Expected outcome:** **Actual negative WaitGroup counter** (panic in Go; recovered in test) — `A05`.

```go
// A05: Actual negative WaitGroup
func TestA05_NegativeWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
	defer func() { _ = recover() }() // wg counter going negative panics
	wg.Done()
}
```


### `TestA06_UnlockNotLockedMutex`
- **Scenario:** Unlock a mutex that is not locked.
- **Expected outcome:** **Actual unlock of not locked mutex** (panic in Go; recovered in test) — `A06`.

```go
// A06: Actual unlock of not locked mutex
func TestA06_UnlockNotLockedMutex(t *testing.T) {
	var m sync.Mutex
	defer func() { _ = recover() }() // unlocking unlocked mutex panics
	m.Unlock()
}
```


### `TestA07_ActualLeak_UnbufferedRecvNoPartner`
- **Scenario:** A goroutine blocks forever on `<-c` on an unbuffered channel with no sender.
- **Expected outcome:** **Actual leak / non-cyclic blocking bug** — `A07` (depending on your mapping this may also fall into leak family).

```go
// A07: Actual Leak (Non-Cyclic blocking bug)  (goroutine blocks forever)
func TestA07_ActualLeak_UnbufferedRecvNoPartner(t *testing.T) {
	c := make(chan struct{})
	go func() {
		<-c // blocks forever
	}()
	time.Sleep(10 * time.Millisecond) // keep trace alive briefly
}
```


### `TestA08_ActualDeadlock_CyclicMutex`
- **Scenario:** Two goroutines acquire two locks in opposite order (AB–BA).
- **Expected outcome:** **Actual cyclic deadlock** — `A08`.

```go
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
```


### `TestA09_ConcurrentReceivesSameChannel`
- **Scenario:** Two goroutines concurrently receive from the same channel; only one send is performed.
- **Expected outcome:** **Concurrent receives on same channel** — `A09`.

```go
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
```


## Possible Misuse / Potential Bugs (`Pxx`)
These tests deliberately create **racy schedules** where the problematic order is possible, but the chosen timing attempts to avoid triggering an actual panic/deadlock in the observed run.


### `TestP01_PossibleSendOnClosed`
- **Scenario:** Send and close race on the same channel; schedule biases send before close.
- **Expected outcome:** **Possible send on closed channel** — `P01`.

```go
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
```


### `TestP02_PossibleRecvOnClosed`
- **Scenario:** Receive and close race; schedule biases receive before close by preloading the buffer.
- **Expected outcome:** **Possible receive on closed channel** — `P02`.

```go
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
```


### `TestP03_PossibleNegativeWaitGroup`
- **Scenario:** Two concurrent `Add(1)` operations and two `Done()` operations; ordering could become negative, but timing biases adds first.
- **Expected outcome:** **Possible negative WaitGroup counter** — `P03`.

```go
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
```


### `TestP04_PossibleUnlockNotLockedMutex`
- **Scenario:** Unlock races with lock; timing biases unlock to happen after lock acquisition, avoiding an actual panic.
- **Expected outcome:** **Possible unlock of not locked mutex** — `P04`.

```go
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
```


### `TestP05_PossibleCyclicDeadlock`
- **Scenario:** Two goroutines lock `a` then `b` vs `b` then `a`, but delayed start avoids an actual deadlock.
- **Expected outcome:** **Possible cyclic deadlock** — `P05`.

```go
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
```

---

## Leak Classifications (`Lxx`)

### `TestL00_Leak_Generic`
- **Scenario:** A goroutine blocks forever in a `select` where no case can proceed.
- **Expected outcome:** **Generic leak** — `L00`.

```go
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
```

### `TestL01_LeakUnbuffered_WithPartner`
- **Scenario:** On an unbuffered channel, a first send/recv pair can complete, but a second send has no matching receiver *in that run*, while a receiver operation exists elsewhere in the trace.
- **Expected outcome:** **Leak on unbuffered channel with possible partner** — `L01`.

```go
// L01: Leak on unbuffered channel with possible partner
func TestL01_LeakUnbuffered_WithPartner(t *testing.T) {
	c := make(chan int)

	go func() { c <- 1 }() // send #1
	go func() { <-c }()    // recv #1 partners with send #1

	go func() { c <- 2 }() // send #2 leaks, but there *exists* a recv op on c (possible partner)
	time.Sleep(10 * time.Millisecond)
}
```


### `TestL02_LeakUnbuffered_WithoutPartner`
- **Scenario:** A goroutine blocks forever receiving from an unbuffered channel with no sender.
- **Expected outcome:** **Leak on unbuffered channel without possible partner** — `L02`.

```go
// L02: Leak on unbuffered channel without possible partner
func TestL02_LeakUnbuffered_WithoutPartner(t *testing.T) {
	c := make(chan int)
	go func() { <-c }() // recv leaks; no sender exists
	time.Sleep(10 * time.Millisecond)
}
```


### `TestL03_LeakBuffered_WithPartner`
- **Scenario:** Buffered channel interactions where a blocked send can exist while some receive operation on the channel also exists.
- **Expected outcome:** **Leak on buffered channel with possible partner** — `L03`.

```go
// L03: Leak on buffered channel with possible partner
func TestL03_LeakBuffered_WithPartner(t *testing.T) {
	c := make(chan int, 1)

	go func() { c <- 1 }() // fills buffer
	go func() { <-c }()    // possible partner exists

	go func() { c <- 2 }() // blocks if buffer is full at that moment => leak with partner
	time.Sleep(10 * time.Millisecond)
}
```


### `TestL04_LeakBuffered_WithoutPartner`
- **Scenario:** Buffered channel receive blocks forever on an empty buffer with no sender.
- **Expected outcome:** **Leak on buffered channel without possible partner** — `L04`.

```go
// L04: Leak on buffered channel without possible partner
func TestL04_LeakBuffered_WithoutPartner(t *testing.T) {
	c := make(chan int, 1)
	go func() { <-c }() // recv on empty buffered channel; no sender => leak
	time.Sleep(10 * time.Millisecond)
}
```


### `TestL05_LeakNilChannel`
- **Scenario:** Send on a nil channel (blocks forever).
- **Expected outcome:** **Leak on nil channel** — `L05`.

```go
// L05: Leak on nil channel
func TestL05_LeakNilChannel(t *testing.T) {
	var c chan int         // nil
	go func() { c <- 1 }() // blocks forever
	time.Sleep(10 * time.Millisecond)
}
```


### `TestL06_LeakSelect_WithPartner`
- **Scenario:** A `select` contains a channel send case, and a corresponding receive exists in the program, but the select case does not complete (blocked schedule).
- **Expected outcome:** **Leak on select with possible partner** — `L06`.

```go
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
```


### `TestL07_LeakSelect_WithoutPartner`
- **Scenario:** A `select` contains a send case on a channel with no receiver anywhere.
- **Expected outcome:** **Leak on select without possible partner** — `L07`.

```go
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
```


### `TestL08_LeakMutex`
- **Scenario:** One goroutine holds a mutex; another goroutine blocks forever trying to lock it.
- **Expected outcome:** **Leak on sync.Mutex** — `L08`.

```go
// L08: Leak on sync.Mutex (lock attempt blocks; last lock exists)
func TestL08_LeakMutex(t *testing.T) {
	var m sync.Mutex
	m.Lock() // last lock
	go func() {
		m.Lock() // leaks here
	}()
	time.Sleep(10 * time.Millisecond)
}
```


### `TestL09_LeakWaitGroup`
- **Scenario:** `Wait()` blocks forever because `Done()` is never called.
- **Expected outcome:** **Leak on sync.WaitGroup** — `L09`.

```go
// L09: Leak on sync.WaitGroup (Wait blocks forever)
func TestL09_LeakWaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Wait() // leaks forever because Done never called
	}()
	time.Sleep(10 * time.Millisecond)
}
```


### `TestL10_LeakCond`
- **Scenario:** A goroutine waits on a condition variable, but no signal/broadcast occurs.
- **Expected outcome:** **Leak on sync.Cond** — `L10`.

```go
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
```

### `TestL11_LeakContext`
- **Scenario:** A goroutine blocks on `<-ctx.Done()` where the cancel function is never called.
- **Expected outcome:** **Leak on context / select on context** — `L11`.

```go
// L11: Leak on channel or select on context (block on ctx.Done that never closes)
func TestL11_LeakContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel // never called

	go func() {
		<-ctx.Done() // leaks forever
	}()
	time.Sleep(10 * time.Millisecond)
}
```