# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD2_3U
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:196
```go
185 ...
186 
187 // MD2-3: Sender with PCS, Receiver inside CS
188 // ------------------------------------------------------------
189 
190 // MD2-3U: Unbuffered Variant
191 func TestMixedDeadlock_MD2_3U(t *testing.T) {
192 	var m sync.Mutex
193 	c := make(chan int)
194 
195 	sender := func() {
196 		m.Lock()           // <-------
197 		time.Sleep(50 * time.Millisecond)
198 		m.Unlock()
199 		c <- 1 // send after PCS
200 	}
201 
202 	receiver := func() {
203 		time.Sleep(50 * time.Millisecond) //let sender complete PCS
204 		m.Lock()
205 		<-c // receive inside CS
206 		m.Unlock()
207 
208 ...
```


###  Channel: Receive
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:205
```go
194 ...
195 
196 		m.Lock()
197 		time.Sleep(50 * time.Millisecond)
198 		m.Unlock()
199 		c <- 1 // send after PCS
200 	}
201 
202 	receiver := func() {
203 		time.Sleep(50 * time.Millisecond) //let sender complete PCS
204 		m.Lock()
205 		<-c // receive inside CS           // <-------
206 		m.Unlock()
207 	}
208 
209 	run2(sender, receiver)
210 }
211 
212 // MD2-3B: Buffered Variant
213 func TestMixedDeadlock_MD2_3B(t *testing.T) {
214 	var m sync.Mutex
215 	c := make(chan int, 1)
216 
217 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:204
```go
193 ...
194 
195 	sender := func() {
196 		m.Lock()
197 		time.Sleep(50 * time.Millisecond)
198 		m.Unlock()
199 		c <- 1 // send after PCS
200 	}
201 
202 	receiver := func() {
203 		time.Sleep(50 * time.Millisecond) //let sender complete PCS
204 		m.Lock()           // <-------
205 		<-c // receive inside CS
206 		m.Unlock()
207 	}
208 
209 	run2(sender, receiver)
210 }
211 
212 // MD2-3B: Buffered Variant
213 func TestMixedDeadlock_MD2_3B(t *testing.T) {
214 	var m sync.Mutex
215 
216 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:199
```go
188 ...
189 
190 // MD2-3U: Unbuffered Variant
191 func TestMixedDeadlock_MD2_3U(t *testing.T) {
192 	var m sync.Mutex
193 	c := make(chan int)
194 
195 	sender := func() {
196 		m.Lock()
197 		time.Sleep(50 * time.Millisecond)
198 		m.Unlock()
199 		c <- 1 // send after PCS           // <-------
200 	}
201 
202 	receiver := func() {
203 		time.Sleep(50 * time.Millisecond) //let sender complete PCS
204 		m.Lock()
205 		<-c // receive inside CS
206 		m.Unlock()
207 	}
208 
209 	run2(sender, receiver)
210 
211 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

