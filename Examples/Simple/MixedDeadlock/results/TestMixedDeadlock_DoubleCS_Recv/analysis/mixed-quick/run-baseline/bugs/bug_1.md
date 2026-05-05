# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_DoubleCS_Recv
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:137
```go
126 ...
127 
128 func TestMixedDeadlock_DoubleCS_Recv(t *testing.T) {
129 	var m sync.Mutex
130 	c := make(chan int, 1)
131 
132 	sender := func() {
133 		m.Lock()
134 		c <- 1
135 		m.Unlock()
136 	}
137            // <-------
138 	receiver := func() {
139 		time.Sleep(50 * time.Millisecond)
140 		m.Lock()
141 		<-c
142 		m.Unlock()
143 
144 		time.Sleep(10 * time.Millisecond)
145 
146 		m.Lock()
147 		time.Sleep(10 * time.Millisecond)
148 
149 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:145
```go
134 ...
135 
136 	}
137 
138 	receiver := func() {
139 		time.Sleep(50 * time.Millisecond)
140 		m.Lock()
141 		<-c
142 		m.Unlock()
143 
144 		time.Sleep(10 * time.Millisecond)
145            // <-------
146 		m.Lock()
147 		time.Sleep(10 * time.Millisecond)
148 		m.Unlock()
149 	}
150 
151 	run2(sender, receiver)
152 }
153 
154 func TestMixedDeadlock_DoubleCS_Recv_2(t *testing.T) {
155 	var m sync.Mutex
156 
157 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:144
```go
133 ...
134 
135 		m.Unlock()
136 	}
137 
138 	receiver := func() {
139 		time.Sleep(50 * time.Millisecond)
140 		m.Lock()
141 		<-c
142 		m.Unlock()
143 
144 		time.Sleep(10 * time.Millisecond)           // <-------
145 
146 		m.Lock()
147 		time.Sleep(10 * time.Millisecond)
148 		m.Unlock()
149 	}
150 
151 	run2(sender, receiver)
152 }
153 
154 func TestMixedDeadlock_DoubleCS_Recv_2(t *testing.T) {
155 
156 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:138
```go
127 ...
128 
129 	var m sync.Mutex
130 	c := make(chan int, 1)
131 
132 	sender := func() {
133 		m.Lock()
134 		c <- 1
135 		m.Unlock()
136 	}
137 
138 	receiver := func() {           // <-------
139 		time.Sleep(50 * time.Millisecond)
140 		m.Lock()
141 		<-c
142 		m.Unlock()
143 
144 		time.Sleep(10 * time.Millisecond)
145 
146 		m.Lock()
147 		time.Sleep(10 * time.Millisecond)
148 		m.Unlock()
149 
150 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

