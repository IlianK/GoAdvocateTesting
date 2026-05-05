# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_DoubleCS_Send_Recv_1
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:191
```go
180 ...
181 
182 	c := make(chan int, 1) // buffered
183 
184 	sender := func() {
185 		m.Lock()
186 		time.Sleep(10 * time.Millisecond) // CS1
187 		m.Unlock()
188 
189 		time.Sleep(10 * time.Millisecond)
190 
191 		m.Lock()           // <-------
192 		c <- 1 // CS2 with send
193 		m.Unlock()
194 	}
195 
196 	receiver := func() {
197 		time.Sleep(50 * time.Millisecond)
198 		m.Lock()
199 		time.Sleep(10 * time.Millisecond) // CS1
200 		m.Unlock()
201 
202 
203 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:205
```go
194 ...
195 
196 	receiver := func() {
197 		time.Sleep(50 * time.Millisecond)
198 		m.Lock()
199 		time.Sleep(10 * time.Millisecond) // CS1
200 		m.Unlock()
201 
202 		time.Sleep(10 * time.Millisecond)
203 
204 		m.Lock()
205 		<-c // CS2 with receive           // <-------
206 		m.Unlock()
207 	}
208 
209 	run2(sender, receiver)
210 }
211 
212 func TestMixedDeadlock_DoubleCS_Send_Recv_2(t *testing.T) {
213 	var m sync.Mutex
214 	c := make(chan int, 1)
215 
216 
217 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:204
```go
193 ...
194 
195 
196 	receiver := func() {
197 		time.Sleep(50 * time.Millisecond)
198 		m.Lock()
199 		time.Sleep(10 * time.Millisecond) // CS1
200 		m.Unlock()
201 
202 		time.Sleep(10 * time.Millisecond)
203 
204 		m.Lock()           // <-------
205 		<-c // CS2 with receive
206 		m.Unlock()
207 	}
208 
209 	run2(sender, receiver)
210 }
211 
212 func TestMixedDeadlock_DoubleCS_Send_Recv_2(t *testing.T) {
213 	var m sync.Mutex
214 	c := make(chan int, 1)
215 
216 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:192
```go
181 ...
182 
183 
184 	sender := func() {
185 		m.Lock()
186 		time.Sleep(10 * time.Millisecond) // CS1
187 		m.Unlock()
188 
189 		time.Sleep(10 * time.Millisecond)
190 
191 		m.Lock()
192 		c <- 1 // CS2 with send           // <-------
193 		m.Unlock()
194 	}
195 
196 	receiver := func() {
197 		time.Sleep(50 * time.Millisecond)
198 		m.Lock()
199 		time.Sleep(10 * time.Millisecond) // CS1
200 		m.Unlock()
201 
202 		time.Sleep(10 * time.Millisecond)
203 
204 ...
```


## Replay
**Replaying was not run**.

