# Channel 

These tests demonstrate behaviors of buffered and unbuffered channels.


## Unbuffered Channel Scenarios

Unbuffered channels are channels where sends and receives are synchronous. A send will block until there is a corresponding receive and vice versa. These tests simulate different behaviors with unbuffered channels.

### TestUnbufferedSendRecv

This test demonstrates a simple send/receive on an unbuffered channel. A value is sent into the channel from a goroutine and received in the main test function.

The send operation blocks until the receive operation happens, ensuring synchronous communication.

```go
func TestUnbufferedSendRecv(t *testing.T) {
    ch := make(chan int)
    go func() { ch <- 1 }()
    _ = <-ch
}
```


### TestUnbufferedLeakNoRecv

This test demonstrates a situation where a value is sent to the channel, but there is no receiver. This leads to a potential deadlock or "leak" situation where the sending goroutine blocks forever.

Since there is no receiver, the goroutine sending to the channel will block indefinitely, causing a potential resource leak or a hanging test.

```go
func TestUnbufferedLeakNoRecv(t *testing.T) {
    ch := make(chan int)
    go func() {
        // no recv
        ch <- 1 // L02
    }()
    time.Sleep(10 * time.Millisecond)
}
```


### TestUnbufferedRecvNoSend

This test demonstrates a scenario where the main function tries to receive from an unbuffered channel without any value being sent into it. This results in a deadlock.

Since no value is sent into the channel, the main function will block indefinitely on the receive operation.

```go
func TestUnbufferedRecvNoSend(t *testing.T) {
    ch := make(chan int)
    // no send
    _ = <-ch // L01 or deadlock
}
```



## Buffered Channel Scenarios
Buffered channels allow for non-blocking sends until the buffer is full. Once the buffer is full, the send operation blocks until there is space available in the buffer. The following tests show how buffered channels behave in different scenarios.

### TestBufferedFillNoRead
This test demonstrates the scenario where values are sent into a buffered channel but there is no reader to receive them. This results in a buffer "leak" as the channel continues to store values.

After filling the buffer (with a capacity of 2), the goroutine attempts to send a third value into the channel. Since there is no receiver, this causes a leak. The test doesn't block immediately but demonstrates potential resource issues if there is no partner to read from the channel.

```go
func TestBufferedFillNoRead(t *testing.T) {
    ch := make(chan int, 2)
    ch <- 1
    ch <- 2 // now full
    go func() {
        // no reader
        ch <- 3 // L04 (buffered leak no partner)
    }()
    time.Sleep(10 * time.Millisecond)
}
```


### TestBufferedDrainSlow
This test simulates a scenario where values are sent to a buffered channel, but the channel is drained slowly, introducing a delay in reading.

The values are sent into the buffered channel, and the main function exits quickly, while the goroutine takes a delay before consuming the values. This shows how the buffered channel can hold values while the reader consumes them slowly.

```go
func TestBufferedDrainSlow(t *testing.T) {
    ch := make(chan int, 2)
    ch <- 1
    ch <- 2
    go func() {
        time.Sleep(10 * time.Millisecond)
        <-ch
        <-ch
    }()
    // main returns quickly
}
```

**Comparison**:
- [Bug Types](./results/comparison_pivot_Bug_Types.csv)

- [Total Time](./results/comparison_pivot_Total_Time_s.csv)
