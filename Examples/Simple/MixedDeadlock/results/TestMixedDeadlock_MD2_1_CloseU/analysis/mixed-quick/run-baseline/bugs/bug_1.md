# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD2_1_CloseU
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:81
```go
70 ...
71 
72 	run2(sender, receiver)
73 }
74 
75 // MD2-1U: Unbuffered Close Variant
76 func TestMixedDeadlock_MD2_1_CloseU(t *testing.T) {
77 	var m sync.Mutex
78 	c := make(chan int) // buffered
79 
80 	closer := func() {
81 		m.Lock()           // <-------
82 		close(c) // close inside CS
83 		m.Unlock()
84 	}
85 
86 	receiver := func() {
87 		time.Sleep(50 * time.Millisecond) // let closer go first
88 		m.Lock()
89 		<-c // receive inside CS
90 		m.Unlock()
91 	}
92 
93 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:89
```go
78 ...
79 
80 	closer := func() {
81 		m.Lock()
82 		close(c) // close inside CS
83 		m.Unlock()
84 	}
85 
86 	receiver := func() {
87 		time.Sleep(50 * time.Millisecond) // let closer go first
88 		m.Lock()
89 		<-c // receive inside CS           // <-------
90 		m.Unlock()
91 	}
92 
93 	run2(closer, receiver)
94 }
95 
96 // MD2-1B: Buffered Close Variant
97 func TestMixedDeadlock_MD2_1_CloseB(t *testing.T) {
98 	var m sync.Mutex
99 	c := make(chan int, 1) // buffered
100 
101 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:88
```go
77 ...
78 
79 
80 	closer := func() {
81 		m.Lock()
82 		close(c) // close inside CS
83 		m.Unlock()
84 	}
85 
86 	receiver := func() {
87 		time.Sleep(50 * time.Millisecond) // let closer go first
88 		m.Lock()           // <-------
89 		<-c // receive inside CS
90 		m.Unlock()
91 	}
92 
93 	run2(closer, receiver)
94 }
95 
96 // MD2-1B: Buffered Close Variant
97 func TestMixedDeadlock_MD2_1_CloseB(t *testing.T) {
98 	var m sync.Mutex
99 
100 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:82
```go
71 ...
72 
73 }
74 
75 // MD2-1U: Unbuffered Close Variant
76 func TestMixedDeadlock_MD2_1_CloseU(t *testing.T) {
77 	var m sync.Mutex
78 	c := make(chan int) // buffered
79 
80 	closer := func() {
81 		m.Lock()
82 		close(c) // close inside CS           // <-------
83 		m.Unlock()
84 	}
85 
86 	receiver := func() {
87 		time.Sleep(50 * time.Millisecond) // let closer go first
88 		m.Lock()
89 		<-c // receive inside CS
90 		m.Unlock()
91 	}
92 
93 
94 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

