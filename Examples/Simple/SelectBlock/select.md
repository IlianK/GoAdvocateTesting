# Select

This folder contains **select-based blocking / leak** scenarios. They are meant to validate that the analyzer reliably reports **leaks caused by `select` statements**, including cases with/without potential partners, buffered-channel select blocking, and context-based select blocking.

## Unbuffered

### `Test_Select_Leak_WithoutPartner_Send`
- **Scenario:** A goroutine executes a `select` with a single `case c <- 1` on an **unbuffered** channel, while **no receiver exists** anywhere.
- **Expected outcome:** The select blocks forever -> **leak on select without possible partner** (commonly `L07`).

```go
func Test_Select_Leak_WithoutPartner_Send(t *testing.T) {
	c := make(chan int) // unbuffered, but key is: no receiver exists

	go func() {
		select {
		case c <- 1: // blocks forever => select leak without partner
		}
	}()

	time.Sleep(50 * time.Millisecond)
}

```


### `Test_Select_Leak_WithPartner_Send`
- **Scenario:** A goroutine executes a `select` with `case c <- 1` on an **unbuffered** channel. A receiver goroutine exists, but is **blocked behind a barrier** and never reaches `<-c`.
- **Expected outcome:** The select blocks forever, but a partner operation exists -> **leak on select with possible partner** (commonly `L06`).

```go
func Test_Select_Leak_WithPartner_Send(t *testing.T) {
	c := make(chan int) // unbuffered
	blockRecv := make(chan struct{})

	// Possible partner exists but is prevented from running.
	go func() {
		<-blockRecv
		<-c
	}()

	go func() {
		select {
		case c <- 1: // blocks (receiver exists but is blocked elsewhere) => select leak with possible partner
		}
	}()

	time.Sleep(50 * time.Millisecond)
}
```


## Buffered

### `Test_Select_Buffered_SendBlocksOnFullBuffer_WithPartner`
- **Scenario:** A buffered channel of capacity 1 is filled (`c <- 1`). A goroutine executes `select { case c <- 2: }`, which blocks because the buffer is full. A receiver exists but is **blocked behind a barrier** and never drains.
- **Expected outcome:** The select blocks forever on the buffered send -> **select-related leak** on a buffered channel (commonly `L06` if a partner exists; otherwise may be classified as a buffered-channel leak depending on implementation).

```go
func Test_Select_Buffered_SendBlocksOnFullBuffer_WithPartner(t *testing.T) {
	c := make(chan int, 1)
	c <- 1 // buffer full

	blockRecv := make(chan struct{})
	go func() {
		<-blockRecv
		<-c // partner exists but never runs
	}()

	go func() {
		select {
		case c <- 2: // blocks forever because buffer is full and nobody drains
		}
	}()

	time.Sleep(50 * time.Millisecond)
}
```

## Context / nil channel

### `Test_Select_Leak_OnContextDone`
- **Scenario:** A goroutine executes a `select` that waits on `<-ctx.Done()` where `ctx` is `context.Background()`. `Done()` is a **nil channel**, so it never fires.
- **Expected outcome:** The select blocks forever -> **leak on channel or select on context** (commonly `L11`).

```go
func Test_Select_Leak_OnContextDone(t *testing.T) {
	ctx := context.Background() // Done() is nil channel

	go func() {
		select {
		case <-ctx.Done(): // blocks forever (nil channel) => context/select leak
		}
	}()

	time.Sleep(50 * time.Millisecond)
}
```