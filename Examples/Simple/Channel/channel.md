# Channel

This folder contains **channel misuse / blocking** scenarios.  
They are meant to validate that the analyzer reliably reports **leaks** and **possible races** in unbuffered and buffered channels.

## Unbuffered
### `Test_Unbuffered_Leak_SendWithoutPartner`
- **Scenario:** A goroutine performs `c <- 1` on an **unbuffered** channel with **no receiver**.
- **Expected outcome:** The send blocks forever -> **leak on unbuffered channel without partner** (commonly `L02`).


```go
func Test_Unbuffered_Leak_SendWithoutPartner(t *testing.T) {
	c := make(chan int) // unbuffered

	go func() {
		c <- 1 // blocks forever (no receiver)
	}()
	time.Sleep(50 * time.Millisecond)
}
```

### `Test_Unbuffered_PossibleSendOnClosed`
- **Scenario:** A sender attempts `c <- 1` while another goroutine closes `c` concurrently.
- **Expected outcome:** Send may happen before close or race with it -> **possible send on closed channel** (commonly `P01`).

```go
func Test_Unbuffered_PossibleSendOnClosed(t *testing.T) {
	c := make(chan int) // unbuffered

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	// Receiver exists so send can succeed in some schedules
	go func() {
		defer wg.Done()
		<-start
		select {
		case <-c:
		case <-time.After(20 * time.Millisecond):
		}
	}()

	// Sender races with close
	go func() {
		defer wg.Done()
		<-start
		// Try to send; if close wins, this would panic
		select {
		case c <- 1:
		case <-time.After(10 * time.Millisecond):
		}
	}()

	close(start)

	// Close happens concurrently with the above operations
	time.Sleep(1 * time.Millisecond)
	close(c)

	wg.Wait()
}
```


## Buffered

### `Test_Buffered_Leak_SendBlocksOnFullBuffer`
- **Scenario:** A buffered channel of capacity 1 is filled (`c <- 1`), then another goroutine attempts a second send (`c <- 2`) while **no receiver drains** the buffer.
- **Expected outcome:** The second send blocks forever -> **leak on buffered channel** (`L03`/`L04` depending on partner).

```go
func Test_Buffered_Leak_SendBlocksOnFullBuffer(t *testing.T) {
	c := make(chan int, 1)

	// fill buffer
	c <- 1

	go func() {
		c <- 2 // blocks forever because buffer is full and nobody drains
	}()

	time.Sleep(50 * time.Millisecond)
}
```

### `Test_Buffered_PossibleRecvOnClosed`
- **Scenario:** A receiver reads a buffered value while another goroutine closes the channel with a slight delay.
- **Expected outcome:** The receive happens before close in this run, but there exists a race with close -> **possible receive on closed channel** (commonly `P02`).
- **Notes:** Initial `c <- 42` ensures the receive can complete without blocking.

```go
func Test_Buffered_PossibleRecvOnClosed(t *testing.T) {
	c := make(chan int, 1)
	c <- 42 // ensure recv before close

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() { // receiver
		defer wg.Done()
		<-start
		_ = <-c // gets value first
	}()

	go func() { // closer (races)
		defer wg.Done()
		<-start
		time.Sleep(1 * time.Millisecond)
		close(c)
	}()

	close(start)
	wg.Wait()
	time.Sleep(10 * time.Millisecond)
}
```
