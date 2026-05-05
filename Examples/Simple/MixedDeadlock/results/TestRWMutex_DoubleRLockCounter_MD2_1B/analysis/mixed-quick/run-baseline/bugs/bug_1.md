# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRWMutex_DoubleRLockCounter_MD2_1B
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:40
```go
29 ...
30 
31 		rw.RUnlock() // Reader Count: 2 -> 1
32 		time.Sleep(30 * time.Millisecond)
33 		// Receive still in CS
34 		<-c
35 
36 		rw.RUnlock()
37 	}
38 
39 	sender := func() {
40 		rw.Lock()           // <-------
41 		c <- 1 // Send in CS
42 		rw.Unlock()
43 	}
44 
45 	run2(reader, sender)
46 }
47 
48 // Multiple Read Locks/Unlocks with Unbuffered Channel
49 func TestRWMutex_DoubleRUnlock_MD2_2U(t *testing.T) {
50 	var rw sync.RWMutex
51 
52 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:34
```go
23 ...
24 
25 	c := make(chan int, 1) // buffered für MD2-1
26 
27 	reader := func() {
28 		time.Sleep(50 * time.Millisecond) // let sender go first
29 		rw.RLock()
30 		rw.RLock()
31 		rw.RUnlock() // Reader Count: 2 -> 1
32 		time.Sleep(30 * time.Millisecond)
33 		// Receive still in CS
34 		<-c           // <-------
35 
36 		rw.RUnlock()
37 	}
38 
39 	sender := func() {
40 		rw.Lock()
41 		c <- 1 // Send in CS
42 		rw.Unlock()
43 	}
44 
45 
46 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:29
```go
18 ...
19 
20 // ------------------------------------------------------------
21 
22 // Multiple Read Locks/Unlocks
23 func TestRWMutex_DoubleRLockCounter_MD2_1B(t *testing.T) {
24 	var rw sync.RWMutex
25 	c := make(chan int, 1) // buffered für MD2-1
26 
27 	reader := func() {
28 		time.Sleep(50 * time.Millisecond) // let sender go first
29 		rw.RLock()           // <-------
30 		rw.RLock()
31 		rw.RUnlock() // Reader Count: 2 -> 1
32 		time.Sleep(30 * time.Millisecond)
33 		// Receive still in CS
34 		<-c
35 
36 		rw.RUnlock()
37 	}
38 
39 	sender := func() {
40 
41 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:41
```go
30 ...
31 
32 		time.Sleep(30 * time.Millisecond)
33 		// Receive still in CS
34 		<-c
35 
36 		rw.RUnlock()
37 	}
38 
39 	sender := func() {
40 		rw.Lock()
41 		c <- 1 // Send in CS           // <-------
42 		rw.Unlock()
43 	}
44 
45 	run2(reader, sender)
46 }
47 
48 // Multiple Read Locks/Unlocks with Unbuffered Channel
49 func TestRWMutex_DoubleRUnlock_MD2_2U(t *testing.T) {
50 	var rw sync.RWMutex
51 	c := make(chan int) // buffered
52 
53 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

