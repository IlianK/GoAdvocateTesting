# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_DoubleCS_Send_2
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:121
```go
110 ...
111 
112 
113 		m.Lock()
114 		c <- 1
115 		m.Unlock()
116 	}
117 
118 	receiver := func() {
119 		time.Sleep(150 * time.Millisecond)
120 		m.Lock()
121 		<-c           // <-------
122 		m.Unlock()
123 	}
124 
125 	run2(sender, receiver)
126 }
127 
128 func TestMixedDeadlock_DoubleCS_Recv(t *testing.T) {
129 	var m sync.Mutex
130 	c := make(chan int, 1)
131 
132 
133 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:120
```go
109 ...
110 
111 		time.Sleep(10 * time.Millisecond)
112 
113 		m.Lock()
114 		c <- 1
115 		m.Unlock()
116 	}
117 
118 	receiver := func() {
119 		time.Sleep(150 * time.Millisecond)
120 		m.Lock()           // <-------
121 		<-c
122 		m.Unlock()
123 	}
124 
125 	run2(sender, receiver)
126 }
127 
128 func TestMixedDeadlock_DoubleCS_Recv(t *testing.T) {
129 	var m sync.Mutex
130 	c := make(chan int, 1)
131 
132 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:114
```go
103 ...
104 
105 
106 	sender := func() {
107 		m.Lock()
108 		time.Sleep(10 * time.Millisecond)
109 		m.Unlock()
110 
111 		time.Sleep(10 * time.Millisecond)
112 
113 		m.Lock()
114 		c <- 1           // <-------
115 		m.Unlock()
116 	}
117 
118 	receiver := func() {
119 		time.Sleep(150 * time.Millisecond)
120 		m.Lock()
121 		<-c
122 		m.Unlock()
123 	}
124 
125 
126 ...
```


###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:113
```go
102 ...
103 
104 	c := make(chan int, 1)
105 
106 	sender := func() {
107 		m.Lock()
108 		time.Sleep(10 * time.Millisecond)
109 		m.Unlock()
110 
111 		time.Sleep(10 * time.Millisecond)
112 
113 		m.Lock()           // <-------
114 		c <- 1
115 		m.Unlock()
116 	}
117 
118 	receiver := func() {
119 		time.Sleep(150 * time.Millisecond)
120 		m.Lock()
121 		<-c
122 		m.Unlock()
123 	}
124 
125 ...
```


## Replay
**Replaying was not run**.

