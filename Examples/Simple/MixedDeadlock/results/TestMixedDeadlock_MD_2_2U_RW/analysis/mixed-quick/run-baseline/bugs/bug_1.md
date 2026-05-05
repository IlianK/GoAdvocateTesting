# Bug: P06 - Possible Mixed Deadlock

The analysis detected a Possible Mixed Deadlock.
A mixed deadlock is a situation, where two routines are blocked on each other, because they are waiting to send or receive on a channel, while holding locks that the other routine needs to proceed.
This can lead to the program getting stuck, if one of the routines is the main routine. Otherwise it can lead to an unnecessary use of resources.

## Test/Program
The bug was found in the following test/program:

- Test/Prog: TestMixedDeadlock_MD_2_2U_RW
- File: /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go
- Trace: advocateTrace_1

## Bug Elements
The elements involved in the found bug are located at the following positions:

###  Mutex: Causing deadlock
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:177
```go
166 ...
167 
168 
169 	writer := func() {
170 		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
171 		rw.Lock()
172 		c <- 1 // send inside CS
173 		rw.Unlock()
174 	}
175 
176 	reader := func() {
177 		rw.RLock()           // <-------
178 		time.Sleep(10 * time.Millisecond)
179 		rw.RUnlock() // PCS
180 		<-c          // receive after PCS
181 	}
182 
183 	run2(reader, writer)
184 }
185 
186 // ------------------------------------------------------------
187 // MD2-3: Sender with PCS, Receiver inside CS
188 
189 ...
```


###  Channel: Send
-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:172
```go
161 ...
162 
163 
164 // MD2-2U: Unbuffered READ/WRTIE Variant
165 func TestMixedDeadlock_MD_2_2U_RW(t *testing.T) {
166 	var rw sync.RWMutex
167 	c := make(chan int)
168 
169 	writer := func() {
170 		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
171 		rw.Lock()
172 		c <- 1 // send inside CS           // <-------
173 		rw.Unlock()
174 	}
175 
176 	reader := func() {
177 		rw.RLock()
178 		time.Sleep(10 * time.Millisecond)
179 		rw.RUnlock() // PCS
180 		<-c          // receive after PCS
181 	}
182 
183 
184 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:171
```go
160 ...
161 
162 }
163 
164 // MD2-2U: Unbuffered READ/WRTIE Variant
165 func TestMixedDeadlock_MD_2_2U_RW(t *testing.T) {
166 	var rw sync.RWMutex
167 	c := make(chan int)
168 
169 	writer := func() {
170 		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
171 		rw.Lock()           // <-------
172 		c <- 1 // send inside CS
173 		rw.Unlock()
174 	}
175 
176 	reader := func() {
177 		rw.RLock()
178 		time.Sleep(10 * time.Millisecond)
179 		rw.RUnlock() // PCS
180 		<-c          // receive after PCS
181 	}
182 
183 ...
```


-> /home/ilian/Projects/ADVOCATE/GoAdvocateTesting/Examples/Simple/MixedDeadlock/TwoCycleMixedDeadlock_test.go:180
```go
169 ...
170 
171 		rw.Lock()
172 		c <- 1 // send inside CS
173 		rw.Unlock()
174 	}
175 
176 	reader := func() {
177 		rw.RLock()
178 		time.Sleep(10 * time.Millisecond)
179 		rw.RUnlock() // PCS
180 		<-c          // receive after PCS           // <-------
181 	}
182 
183 	run2(reader, writer)
184 }
185 
186 // ------------------------------------------------------------
187 // MD2-3: Sender with PCS, Receiver inside CS
188 // ------------------------------------------------------------
189 
190 // MD2-3U: Unbuffered Variant
191 
192 ...
```


## Replay
The bug is a potential bug.
The analyzer has tried to rewrite the trace in such a way that the bug will be triggered when replaying the trace.

**Replaying confirmed the bug**.

It exited with the following code: 42

The replay reached the expected point and found stuck channels.The replay was therefore able to confirm that a mixed deadlock can actually occur.

