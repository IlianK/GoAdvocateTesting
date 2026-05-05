# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD_2_3U_RW
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:291
```go
280 ...
281 
282 
283 	reader := func() {
284 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
285 		rw.RLock()
286 		<-c // receive inside CS
287 		rw.RUnlock()
288 	}
289 
290 	writer := func() {
291 		rw.Lock()           // <-------
292 		time.Sleep(10 * time.Millisecond)
293 		rw.Unlock() // PCS
294 		c <- 1      // send after PCS
295 	}
296 
297 	run2(reader, writer)
298 }
299 
300 // MD2-3B: Buffered READ/WRTIE Variant
301 func TestMixedDeadlock_MD_2_3B_RW(t *testing.T) {
302 
303 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:286
```go
275 ...
276 
277 
278 // MD2-3U: Unbuffered READ/WRTIE Variant
279 func TestMixedDeadlock_MD_2_3U_RW(t *testing.T) {
280 	var rw sync.RWMutex
281 	c := make(chan int)
282 
283 	reader := func() {
284 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
285 		rw.RLock()
286 		<-c // receive inside CS           // <-------
287 		rw.RUnlock()
288 	}
289 
290 	writer := func() {
291 		rw.Lock()
292 		time.Sleep(10 * time.Millisecond)
293 		rw.Unlock() // PCS
294 		c <- 1      // send after PCS
295 	}
296 
297 
298 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:285
```go
274 ...
275 
276 }
277 
278 // MD2-3U: Unbuffered READ/WRTIE Variant
279 func TestMixedDeadlock_MD_2_3U_RW(t *testing.T) {
280 	var rw sync.RWMutex
281 	c := make(chan int)
282 
283 	reader := func() {
284 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
285 		rw.RLock()           // <-------
286 		<-c // receive inside CS
287 		rw.RUnlock()
288 	}
289 
290 	writer := func() {
291 		rw.Lock()
292 		time.Sleep(10 * time.Millisecond)
293 		rw.Unlock() // PCS
294 		c <- 1      // send after PCS
295 	}
296 
297 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:294
```go
283 ...
284 
285 		rw.RLock()
286 		<-c // receive inside CS
287 		rw.RUnlock()
288 	}
289 
290 	writer := func() {
291 		rw.Lock()
292 		time.Sleep(10 * time.Millisecond)
293 		rw.Unlock() // PCS
294 		c <- 1      // send after PCS           // <-------
295 	}
296 
297 	run2(reader, writer)
298 }
299 
300 // MD2-3B: Buffered READ/WRTIE Variant
301 func TestMixedDeadlock_MD_2_3B_RW(t *testing.T) {
302 	var rw sync.RWMutex
303 	c := make(chan int, 1)
304 
305 
306 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

