# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD2_2U
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:155
```go
144 ...
145 
146 
147 	sender := func() {
148 		time.Sleep(50 * time.Millisecond) // let receiver complete PCS
149 		m.Lock()
150 		c <- 1 // send inside CS
151 		m.Unlock()
152 	}
153 
154 	receiver := func() {
155 		m.Lock()           // <-------
156 		time.Sleep(50 * time.Millisecond)
157 		m.Unlock()
158 		<-c // receive after PCS
159 	}
160 
161 	run2(sender, receiver)
162 }
163 
164 // MD2-2U: Unbuffered READ/WRTIE Variant
165 func TestMixedDeadlock_MD_2_2U_RW(t *testing.T) {
166 
167 ...
```


###  Channel: Send
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:150
```go
139 ...
140 
141 
142 // MD2-2U: Unbuffered Variant
143 func TestMixedDeadlock_MD2_2U(t *testing.T) {
144 	var m sync.Mutex
145 	c := make(chan int)
146 
147 	sender := func() {
148 		time.Sleep(50 * time.Millisecond) // let receiver complete PCS
149 		m.Lock()
150 		c <- 1 // send inside CS           // <-------
151 		m.Unlock()
152 	}
153 
154 	receiver := func() {
155 		m.Lock()
156 		time.Sleep(50 * time.Millisecond)
157 		m.Unlock()
158 		<-c // receive after PCS
159 	}
160 
161 
162 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:149
```go
138 ...
139 
140 // ------------------------------------------------------------
141 
142 // MD2-2U: Unbuffered Variant
143 func TestMixedDeadlock_MD2_2U(t *testing.T) {
144 	var m sync.Mutex
145 	c := make(chan int)
146 
147 	sender := func() {
148 		time.Sleep(50 * time.Millisecond) // let receiver complete PCS
149 		m.Lock()           // <-------
150 		c <- 1 // send inside CS
151 		m.Unlock()
152 	}
153 
154 	receiver := func() {
155 		m.Lock()
156 		time.Sleep(50 * time.Millisecond)
157 		m.Unlock()
158 		<-c // receive after PCS
159 	}
160 
161 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:158
```go
147 ...
148 
149 		m.Lock()
150 		c <- 1 // send inside CS
151 		m.Unlock()
152 	}
153 
154 	receiver := func() {
155 		m.Lock()
156 		time.Sleep(50 * time.Millisecond)
157 		m.Unlock()
158 		<-c // receive after PCS           // <-------
159 	}
160 
161 	run2(sender, receiver)
162 }
163 
164 // MD2-2U: Unbuffered READ/WRTIE Variant
165 func TestMixedDeadlock_MD_2_2U_RW(t *testing.T) {
166 	var rw sync.RWMutex
167 	c := make(chan int)
168 
169 
170 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

