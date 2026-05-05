# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MultiDep_FirstOnly
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:320
```go
309 ...
310 
311 
312 		// D4 = (G2, rcv(c), {m})
313 		m.Lock()
314 		<-c
315 		m.Unlock()
316 	}
317 
318 	sender := func() { // G1
319 		// D1 = (G1, snd(c), {m})
320 		m.Lock()           // <-------
321 		c <- 1
322 		m.Unlock()
323 
324 		<-s // sync with G2
325 
326 		// D2 = (G1, snd(c), {m})
327 		m.Lock()
328 		s <- 1
329 		c <- 1
330 		m.Unlock()
331 
332 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:307
```go
296 ...
297 
298 	var m sync.Mutex
299 	c := make(chan int, 1) // buffered
300 	s := make(chan int, 1) // buffered
301 
302 	receiver := func() { // G2
303 		s <- 1 // let G1 continue
304 
305 		// D3 = (G2, rcv(c), {m})
306 		m.Lock()
307 		<-c           // <-------
308 		m.Unlock()
309 
310 		time.Sleep(20 * time.Millisecond) // remove => likely deadlock
311 
312 		// D4 = (G2, rcv(c), {m})
313 		m.Lock()
314 		<-c
315 		m.Unlock()
316 	}
317 
318 
319 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:306
```go
295 ...
296 
297 func TestMixedDeadlock_MultiDep_FirstOnly(t *testing.T) {
298 	var m sync.Mutex
299 	c := make(chan int, 1) // buffered
300 	s := make(chan int, 1) // buffered
301 
302 	receiver := func() { // G2
303 		s <- 1 // let G1 continue
304 
305 		// D3 = (G2, rcv(c), {m})
306 		m.Lock()           // <-------
307 		<-c
308 		m.Unlock()
309 
310 		time.Sleep(20 * time.Millisecond) // remove => likely deadlock
311 
312 		// D4 = (G2, rcv(c), {m})
313 		m.Lock()
314 		<-c
315 		m.Unlock()
316 	}
317 
318 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:321
```go
310 ...
311 
312 		// D4 = (G2, rcv(c), {m})
313 		m.Lock()
314 		<-c
315 		m.Unlock()
316 	}
317 
318 	sender := func() { // G1
319 		// D1 = (G1, snd(c), {m})
320 		m.Lock()
321 		c <- 1           // <-------
322 		m.Unlock()
323 
324 		<-s // sync with G2
325 
326 		// D2 = (G1, snd(c), {m})
327 		m.Lock()
328 		s <- 1
329 		c <- 1
330 		m.Unlock()
331 	}
332 
333 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

