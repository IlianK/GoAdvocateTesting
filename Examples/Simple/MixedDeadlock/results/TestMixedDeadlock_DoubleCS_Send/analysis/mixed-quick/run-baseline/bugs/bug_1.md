# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_DoubleCS_Send
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:84
```go
73 ...
74 
75 func TestMixedDeadlock_DoubleCS_Send(t *testing.T) {
76 	var m sync.Mutex
77 	c := make(chan int, 1)
78 
79 	sender := func() {
80 		m.Lock()
81 		c <- 1
82 		m.Unlock()
83 
84 		time.Sleep(10 * time.Millisecond)           // <-------
85 
86 		m.Lock()
87 		time.Sleep(10 * time.Millisecond)
88 		m.Unlock()
89 
90 	}
91 
92 	receiver := func() {
93 		time.Sleep(50 * time.Millisecond)
94 		m.Lock()
95 
96 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:99
```go
88 ...
89 
90 	}
91 
92 	receiver := func() {
93 		time.Sleep(50 * time.Millisecond)
94 		m.Lock()
95 		<-c
96 		m.Unlock()
97 	}
98 
99 	run2(sender, receiver)           // <-------
100 }
101 
102 func TestMixedDeadlock_DoubleCS_Send_2(t *testing.T) {
103 	var m sync.Mutex
104 	c := make(chan int, 1)
105 
106 	sender := func() {
107 		m.Lock()
108 		time.Sleep(10 * time.Millisecond)
109 		m.Unlock()
110 
111 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:98
```go
87 ...
88 
89 
90 	}
91 
92 	receiver := func() {
93 		time.Sleep(50 * time.Millisecond)
94 		m.Lock()
95 		<-c
96 		m.Unlock()
97 	}
98            // <-------
99 	run2(sender, receiver)
100 }
101 
102 func TestMixedDeadlock_DoubleCS_Send_2(t *testing.T) {
103 	var m sync.Mutex
104 	c := make(chan int, 1)
105 
106 	sender := func() {
107 		m.Lock()
108 		time.Sleep(10 * time.Millisecond)
109 
110 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:85
```go
74 ...
75 
76 	var m sync.Mutex
77 	c := make(chan int, 1)
78 
79 	sender := func() {
80 		m.Lock()
81 		c <- 1
82 		m.Unlock()
83 
84 		time.Sleep(10 * time.Millisecond)
85            // <-------
86 		m.Lock()
87 		time.Sleep(10 * time.Millisecond)
88 		m.Unlock()
89 
90 	}
91 
92 	receiver := func() {
93 		time.Sleep(50 * time.Millisecond)
94 		m.Lock()
95 		<-c
96 
97 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

