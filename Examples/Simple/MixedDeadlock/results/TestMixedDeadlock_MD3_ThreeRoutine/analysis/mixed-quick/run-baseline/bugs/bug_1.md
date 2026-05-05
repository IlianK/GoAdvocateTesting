# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD3_ThreeRoutine
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:342
```go
331 ...
332 
333 	run2(receiver, sender)
334 }
335 
336 func TestMixedDeadlock_MD3_ThreeRoutine(t *testing.T) {
337 	var x, y sync.Mutex
338 	c := make(chan int, 1) // buffered
339 
340 	// R2: hold x, send on c (CS on x), release x
341 	r2 := func() {
342 		x.Lock()           // <-------
343 		c <- 1 // send while holding x CD_R2
344 		x.Unlock()
345 	}
346 
347 	// R3: hold x, receive on c (CS on x), then acquire y
348 	// The sleep ensures R2 sends first in the working trace
349 	r3 := func() {
350 		time.Sleep(60 * time.Millisecond) // let R2 go first
351 		x.Lock()
352 		<-c // receive while holding x CD_R3
353 
354 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:352
```go
341 ...
342 
343 		c <- 1 // send while holding x CD_R2
344 		x.Unlock()
345 	}
346 
347 	// R3: hold x, receive on c (CS on x), then acquire y
348 	// The sleep ensures R2 sends first in the working trace
349 	r3 := func() {
350 		time.Sleep(60 * time.Millisecond) // let R2 go first
351 		x.Lock()
352 		<-c // receive while holding x CD_R3           // <-------
353 		x.Unlock()
354 
355 		y.Lock() // RD_R3: acquire y after releasing x
356 		y.Unlock()
357 	}
358 
359 	// R4: hold y first, then acquire x
360 	// ensures R4 does not interfere with R2/R3 in the working trace.
361 	r4 := func() {
362 		time.Sleep(120 * time.Millisecond) // let R2 and R3 finish first
363 
364 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:351
```go
340 ...
341 
342 		x.Lock()
343 		c <- 1 // send while holding x CD_R2
344 		x.Unlock()
345 	}
346 
347 	// R3: hold x, receive on c (CS on x), then acquire y
348 	// The sleep ensures R2 sends first in the working trace
349 	r3 := func() {
350 		time.Sleep(60 * time.Millisecond) // let R2 go first
351 		x.Lock()           // <-------
352 		<-c // receive while holding x CD_R3
353 		x.Unlock()
354 
355 		y.Lock() // RD_R3: acquire y after releasing x
356 		y.Unlock()
357 	}
358 
359 	// R4: hold y first, then acquire x
360 	// ensures R4 does not interfere with R2/R3 in the working trace.
361 	r4 := func() {
362 
363 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:343
```go
332 ...
333 
334 }
335 
336 func TestMixedDeadlock_MD3_ThreeRoutine(t *testing.T) {
337 	var x, y sync.Mutex
338 	c := make(chan int, 1) // buffered
339 
340 	// R2: hold x, send on c (CS on x), release x
341 	r2 := func() {
342 		x.Lock()
343 		c <- 1 // send while holding x CD_R2           // <-------
344 		x.Unlock()
345 	}
346 
347 	// R3: hold x, receive on c (CS on x), then acquire y
348 	// The sleep ensures R2 sends first in the working trace
349 	r3 := func() {
350 		time.Sleep(60 * time.Millisecond) // let R2 go first
351 		x.Lock()
352 		<-c // receive while holding x CD_R3
353 		x.Unlock()
354 
355 ...
```


## Replay
**Replaying was not run**.

