# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD_2_3CloseU
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:240
```go
229 ...
230 
231 	run2(sender, receiver)
232 }
233 
234 // MD2-3U: Unbuffered Close Variant
235 func TestMixedDeadlock_MD_2_3CloseU(t *testing.T) {
236 	var m sync.Mutex
237 	c := make(chan int) // unbuffered
238 
239 	closer := func() {
240 		m.Lock()           // <-------
241 		time.Sleep(50 * time.Millisecond)
242 		m.Unlock()
243 		close(c) // close after PCS
244 	}
245 
246 	receiver := func() {
247 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
248 		m.Lock()
249 		<-c // receive inside CS
250 		m.Unlock()
251 
252 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:249
```go
238 ...
239 
240 		m.Lock()
241 		time.Sleep(50 * time.Millisecond)
242 		m.Unlock()
243 		close(c) // close after PCS
244 	}
245 
246 	receiver := func() {
247 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
248 		m.Lock()
249 		<-c // receive inside CS           // <-------
250 		m.Unlock()
251 	}
252 
253 	run2(receiver, closer)
254 }
255 
256 // MD2-3B: Buffered Close Variant
257 func TestMixedDeadlock_MD_2_3CloseB(t *testing.T) {
258 	var m sync.Mutex
259 	c := make(chan int, 1) // buffered
260 
261 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:248
```go
237 ...
238 
239 	closer := func() {
240 		m.Lock()
241 		time.Sleep(50 * time.Millisecond)
242 		m.Unlock()
243 		close(c) // close after PCS
244 	}
245 
246 	receiver := func() {
247 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
248 		m.Lock()           // <-------
249 		<-c // receive inside CS
250 		m.Unlock()
251 	}
252 
253 	run2(receiver, closer)
254 }
255 
256 // MD2-3B: Buffered Close Variant
257 func TestMixedDeadlock_MD_2_3CloseB(t *testing.T) {
258 	var m sync.Mutex
259 
260 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:243
```go
232 ...
233 
234 // MD2-3U: Unbuffered Close Variant
235 func TestMixedDeadlock_MD_2_3CloseU(t *testing.T) {
236 	var m sync.Mutex
237 	c := make(chan int) // unbuffered
238 
239 	closer := func() {
240 		m.Lock()
241 		time.Sleep(50 * time.Millisecond)
242 		m.Unlock()
243 		close(c) // close after PCS           // <-------
244 	}
245 
246 	receiver := func() {
247 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
248 		m.Lock()
249 		<-c // receive inside CS
250 		m.Unlock()
251 	}
252 
253 	run2(receiver, closer)
254 
255 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

