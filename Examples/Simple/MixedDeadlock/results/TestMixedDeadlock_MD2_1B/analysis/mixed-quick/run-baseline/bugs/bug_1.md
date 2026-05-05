# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD2_1B
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:59
```go
48 ...
49 
50 	m.Unlock()
51 }
52 
53 // MD2-1B: Buffered Variant
54 func TestMixedDeadlock_MD2_1B(t *testing.T) {
55 	var m sync.Mutex
56 	c := make(chan int, 1) // buffered
57 
58 	sender := func() {
59 		m.Lock()           // <-------
60 		c <- 1 // send inside CS
61 		m.Unlock()
62 
63 	}
64 
65 	receiver := func() {
66 		time.Sleep(50 * time.Millisecond) // let sender go first
67 		m.Lock()
68 		<-c // receive inside CS
69 		m.Unlock()
70 
71 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:68
```go
57 ...
58 
59 		m.Lock()
60 		c <- 1 // send inside CS
61 		m.Unlock()
62 
63 	}
64 
65 	receiver := func() {
66 		time.Sleep(50 * time.Millisecond) // let sender go first
67 		m.Lock()
68 		<-c // receive inside CS           // <-------
69 		m.Unlock()
70 	}
71 
72 	run2(sender, receiver)
73 }
74 
75 // MD2-1U: Unbuffered Close Variant
76 func TestMixedDeadlock_MD2_1_CloseU(t *testing.T) {
77 	var m sync.Mutex
78 	c := make(chan int) // buffered
79 
80 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:67
```go
56 ...
57 
58 	sender := func() {
59 		m.Lock()
60 		c <- 1 // send inside CS
61 		m.Unlock()
62 
63 	}
64 
65 	receiver := func() {
66 		time.Sleep(50 * time.Millisecond) // let sender go first
67 		m.Lock()           // <-------
68 		<-c // receive inside CS
69 		m.Unlock()
70 	}
71 
72 	run2(sender, receiver)
73 }
74 
75 // MD2-1U: Unbuffered Close Variant
76 func TestMixedDeadlock_MD2_1_CloseU(t *testing.T) {
77 	var m sync.Mutex
78 
79 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:60
```go
49 ...
50 
51 }
52 
53 // MD2-1B: Buffered Variant
54 func TestMixedDeadlock_MD2_1B(t *testing.T) {
55 	var m sync.Mutex
56 	c := make(chan int, 1) // buffered
57 
58 	sender := func() {
59 		m.Lock()
60 		c <- 1 // send inside CS           // <-------
61 		m.Unlock()
62 
63 	}
64 
65 	receiver := func() {
66 		time.Sleep(50 * time.Millisecond) // let sender go first
67 		m.Lock()
68 		<-c // receive inside CS
69 		m.Unlock()
70 	}
71 
72 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

