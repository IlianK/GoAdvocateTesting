# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD2_3B
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:227
```go
216 ...
217 
218 		m.Lock()
219 		time.Sleep(10 * time.Millisecond)
220 		m.Unlock()
221 		c <- 1 // send after PCS
222 	}
223 
224 	receiver := func() {
225 		time.Sleep(50 * time.Millisecond) // let sender complete PCS
226 		m.Lock()
227 		<-c // receive inside CS           // <-------
228 		m.Unlock()
229 	}
230 
231 	run2(sender, receiver)
232 }
233 
234 // MD2-3U: Unbuffered Close Variant
235 func TestMixedDeadlock_MD_2_3CloseU(t *testing.T) {
236 	var m sync.Mutex
237 	c := make(chan int) // unbuffered
238 
239 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:226
```go
215 ...
216 
217 	sender := func() {
218 		m.Lock()
219 		time.Sleep(10 * time.Millisecond)
220 		m.Unlock()
221 		c <- 1 // send after PCS
222 	}
223 
224 	receiver := func() {
225 		time.Sleep(50 * time.Millisecond) // let sender complete PCS
226 		m.Lock()           // <-------
227 		<-c // receive inside CS
228 		m.Unlock()
229 	}
230 
231 	run2(sender, receiver)
232 }
233 
234 // MD2-3U: Unbuffered Close Variant
235 func TestMixedDeadlock_MD_2_3CloseU(t *testing.T) {
236 	var m sync.Mutex
237 
238 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:221
```go
210 ...
211 
212 // MD2-3B: Buffered Variant
213 func TestMixedDeadlock_MD2_3B(t *testing.T) {
214 	var m sync.Mutex
215 	c := make(chan int, 1)
216 
217 	sender := func() {
218 		m.Lock()
219 		time.Sleep(10 * time.Millisecond)
220 		m.Unlock()
221 		c <- 1 // send after PCS           // <-------
222 	}
223 
224 	receiver := func() {
225 		time.Sleep(50 * time.Millisecond) // let sender complete PCS
226 		m.Lock()
227 		<-c // receive inside CS
228 		m.Unlock()
229 	}
230 
231 	run2(sender, receiver)
232 
233 ...
```


###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:218
```go
207 ...
208 
209 	run2(sender, receiver)
210 }
211 
212 // MD2-3B: Buffered Variant
213 func TestMixedDeadlock_MD2_3B(t *testing.T) {
214 	var m sync.Mutex
215 	c := make(chan int, 1)
216 
217 	sender := func() {
218 		m.Lock()           // <-------
219 		time.Sleep(10 * time.Millisecond)
220 		m.Unlock()
221 		c <- 1 // send after PCS
222 	}
223 
224 	receiver := func() {
225 		time.Sleep(50 * time.Millisecond) // let sender complete PCS
226 		m.Lock()
227 		<-c // receive inside CS
228 		m.Unlock()
229 
230 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

