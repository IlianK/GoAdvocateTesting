# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD_2_1B_RW
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:131
```go
120 ...
121 
122 	writer := func() {
123 		rw.Lock()
124 		c <- 1 // send inside CS
125 		rw.Unlock()
126 	}
127 
128 	reader := func() {
129 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
130 		rw.RLock()
131 		<-c // receive in CS           // <-------
132 		rw.RUnlock()
133 	}
134 
135 	run2(reader, writer)
136 }
137 
138 // ------------------------------------------------------------
139 // MD2-2: Sender inside CS, Receiver with PCS
140 // ------------------------------------------------------------
141 
142 
143 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:130
```go
119 ...
120 
121 
122 	writer := func() {
123 		rw.Lock()
124 		c <- 1 // send inside CS
125 		rw.Unlock()
126 	}
127 
128 	reader := func() {
129 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
130 		rw.RLock()           // <-------
131 		<-c // receive in CS
132 		rw.RUnlock()
133 	}
134 
135 	run2(reader, writer)
136 }
137 
138 // ------------------------------------------------------------
139 // MD2-2: Sender inside CS, Receiver with PCS
140 // ------------------------------------------------------------
141 
142 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:124
```go
113 ...
114 
115 }
116 
117 // MD2-1B: Buffered READ/WRTIE Variant
118 func TestMixedDeadlock_MD_2_1B_RW(t *testing.T) {
119 	var rw sync.RWMutex
120 	c := make(chan int, 1)
121 
122 	writer := func() {
123 		rw.Lock()
124 		c <- 1 // send inside CS           // <-------
125 		rw.Unlock()
126 	}
127 
128 	reader := func() {
129 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
130 		rw.RLock()
131 		<-c // receive in CS
132 		rw.RUnlock()
133 	}
134 
135 
136 ...
```


###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:123
```go
112 ...
113 
114 	run2(closer, receiver)
115 }
116 
117 // MD2-1B: Buffered READ/WRTIE Variant
118 func TestMixedDeadlock_MD_2_1B_RW(t *testing.T) {
119 	var rw sync.RWMutex
120 	c := make(chan int, 1)
121 
122 	writer := func() {
123 		rw.Lock()           // <-------
124 		c <- 1 // send inside CS
125 		rw.Unlock()
126 	}
127 
128 	reader := func() {
129 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
130 		rw.RLock()
131 		<-c // receive in CS
132 		rw.RUnlock()
133 	}
134 
135 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

