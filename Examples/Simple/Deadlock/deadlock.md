# Deadlocks

This folder contains **cyclic deadlock** scenarios. They are meant to validate that the analyzer reliably reports **cyclic blocking bugs** (commonly `A08`) across different synchronization patterns.

## Mutex cycles

### `Test_CyclicDeadlock_ABBA_TwoMutexes`
- **Scenario:** Two goroutines acquire two mutexes (`a`, `b`) in opposite order (AB–BA).
- **Expected outcome:** Classic cyclic deadlock -> **actual cyclic deadlock** (commonly `A08`).

```go
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
```

### `Test_CyclicDeadlock_ThreeWayMutexCycle`
- **Scenario:** Three goroutines form a 3-way lock cycle (A→B, B→C, C→A).
- **Expected outcome:** Cyclic deadlock involving three locks -> **actual cyclic deadlock** (commonly `A08`).

```go
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
```

## RWMutex cycles

### `Test_CyclicDeadlock_RWMutex_WriterThenReader`
- **Scenario:** One goroutine holds `RLock()` and then attempts to acquire `Lock()` (upgrade attempt), while another goroutine attempts `Lock()` concurrently.
- **Expected outcome:** RWMutex upgrade-style deadlock pattern -> **actual cyclic deadlock** (commonly `A08`).

```go
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
```

## Mixed (Channel + Mutex)

### `Test_CyclicDeadlock_ChannelAndMutex`
- **Scenario:** One goroutine holds a mutex while blocking on receiving from an unbuffered channel. Another goroutine needs the same mutex to perform the send.
- **Expected outcome:** Mixed cyclic dependency between **mutex** and **channel** -> **actual cyclic deadlock** (commonly `A08`).

```go
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
```