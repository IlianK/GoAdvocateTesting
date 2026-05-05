# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD_2_3B_RW
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:313
```go
302 ...
303 
304 
305 	reader := func() {
306 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
307 		rw.RLock()
308 		<-c // receive inside CS
309 		rw.RUnlock()
310 	}
311 
312 	writer := func() {
313 		rw.Lock()           // <-------
314 		time.Sleep(10 * time.Millisecond)
315 		rw.Unlock() // PCS
316 		c <- 1      // send after PCS
317 	}
318 
319 	run2(reader, writer)
320 }
321 
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:308
```go
297 ...
298 
299 
300 // MD2-3B: Buffered READ/WRTIE Variant
301 func TestMixedDeadlock_MD_2_3B_RW(t *testing.T) {
302 	var rw sync.RWMutex
303 	c := make(chan int, 1)
304 
305 	reader := func() {
306 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
307 		rw.RLock()
308 		<-c // receive inside CS           // <-------
309 		rw.RUnlock()
310 	}
311 
312 	writer := func() {
313 		rw.Lock()
314 		time.Sleep(10 * time.Millisecond)
315 		rw.Unlock() // PCS
316 		c <- 1      // send after PCS
317 	}
318 
319 
320 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:307
```go
296 ...
297 
298 }
299 
300 // MD2-3B: Buffered READ/WRTIE Variant
301 func TestMixedDeadlock_MD_2_3B_RW(t *testing.T) {
302 	var rw sync.RWMutex
303 	c := make(chan int, 1)
304 
305 	reader := func() {
306 		time.Sleep(50 * time.Millisecond) // let sender finish PCS
307 		rw.RLock()           // <-------
308 		<-c // receive inside CS
309 		rw.RUnlock()
310 	}
311 
312 	writer := func() {
313 		rw.Lock()
314 		time.Sleep(10 * time.Millisecond)
315 		rw.Unlock() // PCS
316 		c <- 1      // send after PCS
317 	}
318 
319 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:316
```go
305 ...
306 
307 		rw.RLock()
308 		<-c // receive inside CS
309 		rw.RUnlock()
310 	}
311 
312 	writer := func() {
313 		rw.Lock()
314 		time.Sleep(10 * time.Millisecond)
315 		rw.Unlock() // PCS
316 		c <- 1      // send after PCS           // <-------
317 	}
318 
319 	run2(reader, writer)
320 }
321 
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

