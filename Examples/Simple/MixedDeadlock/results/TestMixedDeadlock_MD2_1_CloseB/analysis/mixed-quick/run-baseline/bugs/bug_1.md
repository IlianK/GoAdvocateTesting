# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD2_1_CloseB
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:110
```go
99 ...
100 
101 	closer := func() {
102 		m.Lock()
103 		close(c) // close inside CS
104 		m.Unlock()
105 	}
106 
107 	receiver := func() {
108 		time.Sleep(50 * time.Millisecond) // let closer go first
109 		m.Lock()
110 		<-c // receive inside CS           // <-------
111 		m.Unlock()
112 	}
113 
114 	run2(closer, receiver)
115 }
116 
117 // MD2-1B: Buffered READ/WRTIE Variant
118 func TestMixedDeadlock_MD_2_1B_RW(t *testing.T) {
119 	var rw sync.RWMutex
120 	c := make(chan int, 1)
121 
122 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:109
```go
98 ...
99 
100 
101 	closer := func() {
102 		m.Lock()
103 		close(c) // close inside CS
104 		m.Unlock()
105 	}
106 
107 	receiver := func() {
108 		time.Sleep(50 * time.Millisecond) // let closer go first
109 		m.Lock()           // <-------
110 		<-c // receive inside CS
111 		m.Unlock()
112 	}
113 
114 	run2(closer, receiver)
115 }
116 
117 // MD2-1B: Buffered READ/WRTIE Variant
118 func TestMixedDeadlock_MD_2_1B_RW(t *testing.T) {
119 	var rw sync.RWMutex
120 
121 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:103
```go
92 ...
93 
94 }
95 
96 // MD2-1B: Buffered Close Variant
97 func TestMixedDeadlock_MD2_1_CloseB(t *testing.T) {
98 	var m sync.Mutex
99 	c := make(chan int, 1) // buffered
100 
101 	closer := func() {
102 		m.Lock()
103 		close(c) // close inside CS           // <-------
104 		m.Unlock()
105 	}
106 
107 	receiver := func() {
108 		time.Sleep(50 * time.Millisecond) // let closer go first
109 		m.Lock()
110 		<-c // receive inside CS
111 		m.Unlock()
112 	}
113 
114 
115 ...
```


###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:102
```go
91 ...
92 
93 	run2(closer, receiver)
94 }
95 
96 // MD2-1B: Buffered Close Variant
97 func TestMixedDeadlock_MD2_1_CloseB(t *testing.T) {
98 	var m sync.Mutex
99 	c := make(chan int, 1) // buffered
100 
101 	closer := func() {
102 		m.Lock()           // <-------
103 		close(c) // close inside CS
104 		m.Unlock()
105 	}
106 
107 	receiver := func() {
108 		time.Sleep(50 * time.Millisecond) // let closer go first
109 		m.Lock()
110 		<-c // receive inside CS
111 		m.Unlock()
112 	}
113 
114 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

