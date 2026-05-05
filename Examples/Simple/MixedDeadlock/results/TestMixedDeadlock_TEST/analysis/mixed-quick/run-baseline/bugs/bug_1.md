# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_TEST
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:42
```go
31 ...
32 
33 // ------------------------------------------------------------
34 // MD2-1: Both sender and receiver in CS
35 // ------------------------------------------------------------
36 
37 func TestMixedDeadlock_TEST(t *testing.T) {
38 	var m sync.Mutex
39 	c := make(chan int, 1) // buffered
40 
41 	go func() {
42 		m.Lock()           // <-------
43 		c <- 1 // send inside CS
44 		m.Unlock()
45 	}()
46 
47 	time.Sleep(50 * time.Millisecond) // let sender go first
48 	m.Lock()
49 	<-c // receive inside CS
50 	m.Unlock()
51 }
52 
53 
54 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:49
```go
38 ...
39 
40 
41 	go func() {
42 		m.Lock()
43 		c <- 1 // send inside CS
44 		m.Unlock()
45 	}()
46 
47 	time.Sleep(50 * time.Millisecond) // let sender go first
48 	m.Lock()
49 	<-c // receive inside CS           // <-------
50 	m.Unlock()
51 }
52 
53 // MD2-1B: Buffered Variant
54 func TestMixedDeadlock_MD2_1B(t *testing.T) {
55 	var m sync.Mutex
56 	c := make(chan int, 1) // buffered
57 
58 	sender := func() {
59 		m.Lock()
60 
61 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:48
```go
37 ...
38 
39 	c := make(chan int, 1) // buffered
40 
41 	go func() {
42 		m.Lock()
43 		c <- 1 // send inside CS
44 		m.Unlock()
45 	}()
46 
47 	time.Sleep(50 * time.Millisecond) // let sender go first
48 	m.Lock()           // <-------
49 	<-c // receive inside CS
50 	m.Unlock()
51 }
52 
53 // MD2-1B: Buffered Variant
54 func TestMixedDeadlock_MD2_1B(t *testing.T) {
55 	var m sync.Mutex
56 	c := make(chan int, 1) // buffered
57 
58 	sender := func() {
59 
60 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:43
```go
32 ...
33 
34 // MD2-1: Both sender and receiver in CS
35 // ------------------------------------------------------------
36 
37 func TestMixedDeadlock_TEST(t *testing.T) {
38 	var m sync.Mutex
39 	c := make(chan int, 1) // buffered
40 
41 	go func() {
42 		m.Lock()
43 		c <- 1 // send inside CS           // <-------
44 		m.Unlock()
45 	}()
46 
47 	time.Sleep(50 * time.Millisecond) // let sender go first
48 	m.Lock()
49 	<-c // receive inside CS
50 	m.Unlock()
51 }
52 
53 // MD2-1B: Buffered Variant
54 
55 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

