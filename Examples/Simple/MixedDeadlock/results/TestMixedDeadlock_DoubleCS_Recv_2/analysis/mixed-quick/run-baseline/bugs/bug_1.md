# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_DoubleCS_Recv_2
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:159
```go
148 ...
149 
150 
151 	run2(sender, receiver)
152 }
153 
154 func TestMixedDeadlock_DoubleCS_Recv_2(t *testing.T) {
155 	var m sync.Mutex
156 	c := make(chan int, 1)
157 
158 	sender := func() {
159 		m.Lock()           // <-------
160 		c <- 1
161 		m.Unlock()
162 	}
163 
164 	receiver := func() {
165 		time.Sleep(50 * time.Millisecond)
166 		m.Lock()
167 		time.Sleep(10 * time.Millisecond)
168 		m.Unlock()
169 
170 
171 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:173
```go
162 ...
163 
164 	receiver := func() {
165 		time.Sleep(50 * time.Millisecond)
166 		m.Lock()
167 		time.Sleep(10 * time.Millisecond)
168 		m.Unlock()
169 
170 		time.Sleep(10 * time.Millisecond)
171 
172 		m.Lock()
173 		<-c           // <-------
174 		m.Unlock()
175 	}
176 
177 	run2(sender, receiver)
178 }
179 
180 func TestMixedDeadlock_DoubleCS_Send_Recv_1(t *testing.T) {
181 	var m sync.Mutex
182 	c := make(chan int, 1) // buffered
183 
184 
185 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:172
```go
161 ...
162 
163 
164 	receiver := func() {
165 		time.Sleep(50 * time.Millisecond)
166 		m.Lock()
167 		time.Sleep(10 * time.Millisecond)
168 		m.Unlock()
169 
170 		time.Sleep(10 * time.Millisecond)
171 
172 		m.Lock()           // <-------
173 		<-c
174 		m.Unlock()
175 	}
176 
177 	run2(sender, receiver)
178 }
179 
180 func TestMixedDeadlock_DoubleCS_Send_Recv_1(t *testing.T) {
181 	var m sync.Mutex
182 	c := make(chan int, 1) // buffered
183 
184 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:160
```go
149 ...
150 
151 	run2(sender, receiver)
152 }
153 
154 func TestMixedDeadlock_DoubleCS_Recv_2(t *testing.T) {
155 	var m sync.Mutex
156 	c := make(chan int, 1)
157 
158 	sender := func() {
159 		m.Lock()
160 		c <- 1           // <-------
161 		m.Unlock()
162 	}
163 
164 	receiver := func() {
165 		time.Sleep(50 * time.Millisecond)
166 		m.Lock()
167 		time.Sleep(10 * time.Millisecond)
168 		m.Unlock()
169 
170 		time.Sleep(10 * time.Millisecond)
171 
172 ...
```


## Replay
**Replaying was not run**.

