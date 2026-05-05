# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MultiDep_LastOnly
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:281
```go
270 ...
271 
272 	sender := func() { // G1
273 		// D1 = (G1, snd(c), {m})
274 		m.Lock()
275 		c <- 1
276 		m.Unlock()
277 
278 		<-s // sync with G2
279 
280 		// D2 = (G1, snd(c), {m})
281 		m.Lock()           // <-------
282 		s <- 1
283 		c <- 1
284 		m.Unlock()
285 	}
286 
287 	run2(receiver, sender)
288 }
289 
290 /*
291 Cyclic dependency between D2 and D3, and between D2 and D4.
292 
293 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:268
```go
257 ...
258 
259 		m.Lock()
260 		<-c
261 		m.Unlock()
262 
263 		s <- 1 // allow G1 to proceed
264 		<-s    // handshake
265 
266 		// D4 = (G2, rcv(c), {m})
267 		m.Lock()
268 		<-c           // <-------
269 		m.Unlock()
270 	}
271 
272 	sender := func() { // G1
273 		// D1 = (G1, snd(c), {m})
274 		m.Lock()
275 		c <- 1
276 		m.Unlock()
277 
278 		<-s // sync with G2
279 
280 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:267
```go
256 ...
257 
258 		// D3 = (G2, rcv(c), {m})
259 		m.Lock()
260 		<-c
261 		m.Unlock()
262 
263 		s <- 1 // allow G1 to proceed
264 		<-s    // handshake
265 
266 		// D4 = (G2, rcv(c), {m})
267 		m.Lock()           // <-------
268 		<-c
269 		m.Unlock()
270 	}
271 
272 	sender := func() { // G1
273 		// D1 = (G1, snd(c), {m})
274 		m.Lock()
275 		c <- 1
276 		m.Unlock()
277 
278 
279 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:283
```go
272 ...
273 
274 		m.Lock()
275 		c <- 1
276 		m.Unlock()
277 
278 		<-s // sync with G2
279 
280 		// D2 = (G1, snd(c), {m})
281 		m.Lock()
282 		s <- 1
283 		c <- 1           // <-------
284 		m.Unlock()
285 	}
286 
287 	run2(receiver, sender)
288 }
289 
290 /*
291 Cyclic dependency between D2 and D3, and between D2 and D4.
292 But D1 is not involved in any cyclic dependency.
293 
294 
295 ...
```


## Replay
**Replaying was not run**.

