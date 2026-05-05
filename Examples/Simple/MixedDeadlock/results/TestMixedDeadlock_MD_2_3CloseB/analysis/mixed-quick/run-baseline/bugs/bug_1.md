# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD_2_3CloseB
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:262
```go
251 ...
252 
253 	run2(receiver, closer)
254 }
255 
256 // MD2-3B: Buffered Close Variant
257 func TestMixedDeadlock_MD_2_3CloseB(t *testing.T) {
258 	var m sync.Mutex
259 	c := make(chan int, 1) // buffered
260 
261 	closer := func() {
262 		m.Lock()           // <-------
263 		time.Sleep(50 * time.Millisecond)
264 		m.Unlock()
265 		close(c) // close after PCS
266 	}
267 
268 	receiver := func() {
269 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
270 		m.Lock()
271 		<-c // receive inside CS
272 		m.Unlock()
273 
274 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:271
```go
260 ...
261 
262 		m.Lock()
263 		time.Sleep(50 * time.Millisecond)
264 		m.Unlock()
265 		close(c) // close after PCS
266 	}
267 
268 	receiver := func() {
269 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
270 		m.Lock()
271 		<-c // receive inside CS           // <-------
272 		m.Unlock()
273 	}
274 
275 	run2(receiver, closer)
276 }
277 
278 // MD2-3U: Unbuffered READ/WRTIE Variant
279 func TestMixedDeadlock_MD_2_3U_RW(t *testing.T) {
280 	var rw sync.RWMutex
281 	c := make(chan int)
282 
283 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:270
```go
259 ...
260 
261 	closer := func() {
262 		m.Lock()
263 		time.Sleep(50 * time.Millisecond)
264 		m.Unlock()
265 		close(c) // close after PCS
266 	}
267 
268 	receiver := func() {
269 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
270 		m.Lock()           // <-------
271 		<-c // receive inside CS
272 		m.Unlock()
273 	}
274 
275 	run2(receiver, closer)
276 }
277 
278 // MD2-3U: Unbuffered READ/WRTIE Variant
279 func TestMixedDeadlock_MD_2_3U_RW(t *testing.T) {
280 	var rw sync.RWMutex
281 
282 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:265
```go
254 ...
255 
256 // MD2-3B: Buffered Close Variant
257 func TestMixedDeadlock_MD_2_3CloseB(t *testing.T) {
258 	var m sync.Mutex
259 	c := make(chan int, 1) // buffered
260 
261 	closer := func() {
262 		m.Lock()
263 		time.Sleep(50 * time.Millisecond)
264 		m.Unlock()
265 		close(c) // close after PCS           // <-------
266 	}
267 
268 	receiver := func() {
269 		time.Sleep(50 * time.Millisecond) // let closer complete PCS
270 		m.Lock()
271 		<-c // receive inside CS
272 		m.Unlock()
273 	}
274 
275 	run2(receiver, closer)
276 
277 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

