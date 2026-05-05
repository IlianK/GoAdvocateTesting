# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_DoubleCS_Send_Recv_2
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:217
```go
206 ...
207 
208 
209 	run2(sender, receiver)
210 }
211 
212 func TestMixedDeadlock_DoubleCS_Send_Recv_2(t *testing.T) {
213 	var m sync.Mutex
214 	c := make(chan int, 1)
215 
216 	sender := func() {
217 		m.Lock()           // <-------
218 		c <- 1
219 		m.Unlock()
220 
221 		time.Sleep(10 * time.Millisecond)
222 
223 		m.Lock()
224 		time.Sleep(10 * time.Millisecond)
225 		m.Unlock()
226 	}
227 
228 
229 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:231
```go
220 ...
221 
222 
223 		m.Lock()
224 		time.Sleep(10 * time.Millisecond)
225 		m.Unlock()
226 	}
227 
228 	receiver := func() {
229 		time.Sleep(50 * time.Millisecond)
230 		m.Lock()
231 		<-c           // <-------
232 		m.Unlock()
233 
234 		time.Sleep(10 * time.Millisecond)
235 
236 		m.Lock()
237 		time.Sleep(10 * time.Millisecond)
238 		m.Unlock()
239 	}
240 
241 	run2(sender, receiver)
242 
243 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:230
```go
219 ...
220 
221 		time.Sleep(10 * time.Millisecond)
222 
223 		m.Lock()
224 		time.Sleep(10 * time.Millisecond)
225 		m.Unlock()
226 	}
227 
228 	receiver := func() {
229 		time.Sleep(50 * time.Millisecond)
230 		m.Lock()           // <-------
231 		<-c
232 		m.Unlock()
233 
234 		time.Sleep(10 * time.Millisecond)
235 
236 		m.Lock()
237 		time.Sleep(10 * time.Millisecond)
238 		m.Unlock()
239 	}
240 
241 
242 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/GeneralMixedDeadlock_test.go:218
```go
207 ...
208 
209 	run2(sender, receiver)
210 }
211 
212 func TestMixedDeadlock_DoubleCS_Send_Recv_2(t *testing.T) {
213 	var m sync.Mutex
214 	c := make(chan int, 1)
215 
216 	sender := func() {
217 		m.Lock()
218 		c <- 1           // <-------
219 		m.Unlock()
220 
221 		time.Sleep(10 * time.Millisecond)
222 
223 		m.Lock()
224 		time.Sleep(10 * time.Millisecond)
225 		m.Unlock()
226 	}
227 
228 	receiver := func() {
229 
230 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

