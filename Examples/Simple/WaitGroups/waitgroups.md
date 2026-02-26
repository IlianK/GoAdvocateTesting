# WaitGroup

This folder contains **WaitGroup misuse and blocking** scenarios. They are meant to validate that the analyzer reliably reports **actual misuse** (panic-class bugs) and **leaks** involving `sync.WaitGroup`, including common ordering mistakes and situations where `Wait()` never completes.

## Actual misuse

### `Test_WG_A05_NegativeCounter`
- **Scenario:** A `WaitGroup` counter is incremented with `Add(1)` but `Done()` is called **twice**, making the counter negative.
- **Expected outcome:** **Actual negative WaitGroup counter** (panic in Go; recovered in test) — commonly `A05`.

```go
func Test_WG_A05_NegativeCounter(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()

	defer func() { _ = recover() }() // keep test suite running
	wg.Done()                        // counter goes negative => panic
}
```

## Leaks / blocking waits

### `Test_WG_L09_WaitBlocksForever`
- **Scenario:** A goroutine calls `wg.Wait()` after `Add(1)` but **no goroutine ever calls `Done()`**.
- **Expected outcome:** `Wait()` blocks forever -> **leak on sync.WaitGroup** (commonly `L09`).

```go
func Test_WG_L09_WaitBlocksForever(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Wait() // leaks forever because Done never called
	}()

	time.Sleep(50 * time.Millisecond) // allow trace capture
}
```

### `Test_WG_Leak_WorkerNeverDone_BlockedOnChan`
- **Scenario:** A worker goroutine has `defer wg.Done()` but blocks forever on a channel receive, so `Done()` is never executed. Another goroutine waits on `wg.Wait()`.
- **Expected outcome:** `Wait()` blocks forever due to a stuck worker -> **leak on sync.WaitGroup** (commonly `L09`), with the root cause being a blocked worker.

```go
func Test_WG_Possible_AddAfterWait(t *testing.T) {
	var wg sync.WaitGroup

	startWait := make(chan struct{})
	go func() {
		close(startWait)
		wg.Wait() // may return immediately, but Add happens concurrently afterwards (misuse pattern)
	}()

	<-startWait
	time.Sleep(1 * time.Millisecond)
	wg.Add(1)
	// We purposely never call Done to keep the trace interesting; you can add Done to avoid leaking.
	time.Sleep(50 * time.Millisecond)
}
```


## Possible misuse patterns

### `Test_WG_Possible_AddAfterWait`
- **Scenario:** `wg.Wait()` is started in one goroutine while another goroutine calls `wg.Add(1)` **after** the wait begins.
- **Expected outcome:** This is a common misuse pattern that can lead to races or incorrect synchronization; depending on the analyzer, it may be reported as a **possible WaitGroup misuse** (or as a leak if the added work is never completed).

```go
func Test_WG_Leak_WorkerNeverDone_BlockedOnChan(t *testing.T) {
	var wg sync.WaitGroup
	block := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-block // blocks forever, so Done never runs
	}()

	go func() {
		wg.Wait() // leaks forever
	}()

	time.Sleep(50 * time.Millisecond)
}
```