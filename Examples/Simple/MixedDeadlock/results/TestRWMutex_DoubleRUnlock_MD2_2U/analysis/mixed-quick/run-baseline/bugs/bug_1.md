# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestRWMutex_DoubleRUnlock_MD2_2U
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:54
```go
43 ...
44 
45 	run2(reader, sender)
46 }
47 
48 // Multiple Read Locks/Unlocks with Unbuffered Channel
49 func TestRWMutex_DoubleRUnlock_MD2_2U(t *testing.T) {
50 	var rw sync.RWMutex
51 	c := make(chan int) // buffered
52 
53 	reader := func() {
54 		rw.RLock()           // <-------
55 		rw.RLock()
56 		rw.RUnlock()
57 		rw.RUnlock()
58 		<-c // after PCS
59 	}
60 
61 	sender := func() {
62 		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
63 		rw.Lock()
64 		c <- 1 // in CS
65 
66 ...
```


###  Channel: Send
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:64
```go
53 ...
54 
55 		rw.RLock()
56 		rw.RUnlock()
57 		rw.RUnlock()
58 		<-c // after PCS
59 	}
60 
61 	sender := func() {
62 		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
63 		rw.Lock()
64 		c <- 1 // in CS           // <-------
65 		rw.Unlock()
66 	}
67 
68 	run2(reader, sender)
69 }
70 
71 // ------------------------------------------------------------
72 // MULTIPLE CS
73 // ------------------------------------------------------------
74 
75 
76 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:63
```go
52 ...
53 
54 		rw.RLock()
55 		rw.RLock()
56 		rw.RUnlock()
57 		rw.RUnlock()
58 		<-c // after PCS
59 	}
60 
61 	sender := func() {
62 		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
63 		rw.Lock()           // <-------
64 		c <- 1 // in CS
65 		rw.Unlock()
66 	}
67 
68 	run2(reader, sender)
69 }
70 
71 // ------------------------------------------------------------
72 // MULTIPLE CS
73 // ------------------------------------------------------------
74 
75 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:58
```go
47 ...
48 
49 func TestRWMutex_DoubleRUnlock_MD2_2U(t *testing.T) {
50 	var rw sync.RWMutex
51 	c := make(chan int) // buffered
52 
53 	reader := func() {
54 		rw.RLock()
55 		rw.RLock()
56 		rw.RUnlock()
57 		rw.RUnlock()
58 		<-c // after PCS           // <-------
59 	}
60 
61 	sender := func() {
62 		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
63 		rw.Lock()
64 		c <- 1 // in CS
65 		rw.Unlock()
66 	}
67 
68 	run2(reader, sender)
69 
70 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

